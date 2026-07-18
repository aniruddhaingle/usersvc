package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"usersvc/internal/domain"
	"usersvc/internal/repository"
	"usersvc/internal/service"
)

type stubRepo struct {
	user      *domain.User
	createErr error
}

func (s *stubRepo) Create(ctx context.Context, u *domain.User) error { return s.createErr }

func (s *stubRepo) GetUserById(ctx context.Context, id string) (*domain.User, error) {
	return s.user, nil
}

func (s *stubRepo) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return []*domain.User{}, nil
}

func (s *stubRepo) DeleteUser(ctx context.Context, id string) error { return nil }

type stubCache struct{}

func (stubCache) Set(ctx context.Context, key string, value any, exp time.Duration) error {
	return nil
}

func (stubCache) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (stubCache) Del(ctx context.Context, key string) error { return nil }

type stubPublisher struct{}

func (stubPublisher) Publish(ctx context.Context, body []byte) error { return nil }

func newTestMux(repo *stubRepo) *http.ServeMux {
	svc := service.NewUserService(repo, stubCache{}, stubPublisher{})
	h := NewHandler(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func TestGetUserUnknownIDReturns404(t *testing.T) {
	mux := newTestMux(&stubRepo{user: nil})
	req := httptest.NewRequest(http.MethodGet, "/users/does-not-exist", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCreateUserDuplicateEmailReturns409(t *testing.T) {
	repoErr := fmt.Errorf("couldnt create user: %w", repository.ErrDuplicateEmail)
	mux := newTestMux(&stubRepo{createErr: repoErr})
	body := strings.NewReader(`{"email":"a@b.c","password":"pw"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusConflict)
	}
}
