package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"usersvc/internal/domain"
	"usersvc/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/singleflight"
)

type UserRepo interface {
	Create(ctx context.Context, user *domain.User) error
	GetUserById(ctx context.Context, id string) (*domain.User, error)
	ListUsers(ctx context.Context) ([]*domain.User, error)
	DeleteUser(ctx context.Context, id string) error
}

type Cache interface {
	Set(ctx context.Context, key string, value any, exp time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

type EventPublisher interface {
	Publish(ctx context.Context, body []byte) error
}

type UserService struct {
	repo     UserRepo
	redis    Cache
	rabbitmq EventPublisher
	sf       singleflight.Group
}

func NewUserService(repo UserRepo, redis Cache, rabbitmq EventPublisher) *UserService {
	return &UserService{
		repo: repo, redis: redis, rabbitmq: rabbitmq,
	}
}

// say the rabbitmq crashed or smth wrong happens with it, it bare minimum the user creation surely takes place if th required contraints required for it s creations are satisfied
func (u *UserService) Register(ctx context.Context, email string, passwd string) (*domain.User, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("passwd hash couldn't be calcd: %w", err)
	}

	usr := &domain.User{ID: uuid.New().String(), Email: email, PasswordHash: string(hashedPwd), CreatedAt: time.Now()}

	if err := u.repo.Create(ctx, usr); err != nil {
		return nil, fmt.Errorf("error creating user %s: %w", email, err)
	}

	eve := map[string]string{
		"user_id": usr.ID,
		"email":   usr.Email,
	}
	prsd, _ := json.Marshal(eve)
	if err := u.rabbitmq.Publish(ctx, prsd); err != nil {
		return usr, fmt.Errorf("user creation success but event publishment didnt go as intended: %w", err)
	}
	return usr, nil
}

// try for a cache hit
// if cache missed, go for a db fetch, but but but
// there is a chance that a cache-stampede may happen
// to avoid that, we are using SingleFlight here
func (u *UserService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	ckey := "user:" + userID
	dataFound, err := u.redis.Get(ctx, ckey)
	if err != nil {
		slog.Warn("cache read failed", "key", ckey, "error", err)
	}
	if err == nil && dataFound != "" {
		var user domain.User
		if err := json.Unmarshal([]byte(dataFound), &user); err == nil {
			return &user, nil
		}
	}

	val, err, _ := u.sf.Do(ckey, func() (any, error) {
		user, err := u.repo.GetUserById(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, fmt.Errorf("get user %s: %w", userID, repository.ErrUserNotFound)
		}
		//set the cache for future requests
		prsd, _ := json.Marshal(user)
		if err := u.redis.Set(ctx, ckey, string(prsd), time.Minute*10); err != nil {
			slog.Warn("cache set failed", "key", ckey, "error", err)
		}
		return user, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(*domain.User), nil
}

func (u *UserService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return u.repo.ListUsers(ctx)
}

func (u *UserService) DeleteUser(ctx context.Context, userID string) error {
	if err := u.repo.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("error deleting user %s: %w", userID, err)
	}

	err := u.redis.Del(ctx, "user:"+userID)
	if err != nil {
		return fmt.Errorf("error delting user %s from cache: %w", userID, err)
	}
	return nil
}
