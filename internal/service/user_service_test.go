package service

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"usersvc/internal/domain"
)

type fakeRepo struct {
	getCalls int32
	user     *domain.User
	getDelay time.Duration
}

func (f *fakeRepo) Create(ctx context.Context, user *domain.User) error { return nil }

func (f *fakeRepo) GetUserById(ctx context.Context, id string) (*domain.User, error) {
	atomic.AddInt32(&f.getCalls, 1)
	if f.getDelay > 0 {
		time.Sleep(f.getDelay)
	}
	return f.user, nil
}

func (f *fakeRepo) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return []*domain.User{}, nil
}

func (f *fakeRepo) DeleteUser(ctx context.Context, id string) error { return nil }

type fakeCache struct {
	mu    sync.Mutex
	store map[string]string
}

func newFakeCache() *fakeCache { return &fakeCache{store: map[string]string{}} }

func (f *fakeCache) Set(ctx context.Context, key string, value any, exp time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store[key] = value.(string)
	return nil
}

func (f *fakeCache) Get(ctx context.Context, key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v := f.store[key]
	return v, nil
}

func (f *fakeCache) Del(ctx context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.store, key)
	return nil
}

type fakePublisher struct{}

func (fakePublisher) Publish(ctx context.Context, body []byte) error { return nil }

func TestGetUserReturnsCachedValueWithoutRepo(t *testing.T) {
	usr := &domain.User{ID: "id1", Email: "a@b.c", CreatedAt: time.Now().UTC()}
	cache := newFakeCache()
	b, err := json.Marshal(usr)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	cache.store["user:id1"] = string(b)
	repo := &fakeRepo{}
	svc := NewUserService(repo, cache, fakePublisher{})

	got, err := svc.GetUser(context.Background(), "id1")
	if err != nil {
		t.Fatalf("GetUser returned error: %v", err)
	}
	if got.ID != usr.ID || got.Email != usr.Email {
		t.Errorf("got user %+v, want %+v", got, usr)
	}
	if calls := atomic.LoadInt32(&repo.getCalls); calls != 0 {
		t.Errorf("repo was hit %d times, want 0 (cache should have served it)", calls)
	}
}

func TestGetUserPopulatesCacheOnMiss(t *testing.T) {
	usr := &domain.User{ID: "id1", Email: "a@b.c", CreatedAt: time.Now().UTC()}
	cache := newFakeCache()
	repo := &fakeRepo{user: usr}
	svc := NewUserService(repo, cache, fakePublisher{})

	got, err := svc.GetUser(context.Background(), "id1")
	if err != nil {
		t.Fatalf("GetUser returned error: %v", err)
	}
	if got.ID != usr.ID {
		t.Errorf("got user id %q, want %q", got.ID, usr.ID)
	}
	if calls := atomic.LoadInt32(&repo.getCalls); calls != 1 {
		t.Errorf("repo was hit %d times, want 1", calls)
	}
	cached, err := cache.Get(context.Background(), "user:id1")
	if err != nil {
		t.Fatalf("cache was not populated after miss: %v", err)
	}
	var fromCache domain.User
	if err := json.Unmarshal([]byte(cached), &fromCache); err != nil {
		t.Fatalf("cached value is not valid json: %v", err)
	}
	if fromCache.ID != usr.ID {
		t.Errorf("cached user id %q, want %q", fromCache.ID, usr.ID)
	}
}

func TestGetUserSingleflightCollapsesConcurrentMisses(t *testing.T) {
	usr := &domain.User{ID: "id1", Email: "a@b.c", CreatedAt: time.Now().UTC()}
	// the delay holds the first flight open long enough for every
	// concurrent caller to join it instead of starting a new one
	repo := &fakeRepo{user: usr, getDelay: 250 * time.Millisecond}
	cache := newFakeCache()
	svc := NewUserService(repo, cache, fakePublisher{})

	const n = 25
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := svc.GetUser(context.Background(), "id1"); err != nil {
				t.Errorf("GetUser returned error: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if calls := atomic.LoadInt32(&repo.getCalls); calls != 1 {
		t.Errorf("repo called %d times for %d concurrent misses, want exactly 1", calls, n)
	}
}
