package repository

import (
	"adv-bknd/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id,email,password_hash,created_at) VALUES ($1,$2,$3,$4)`
	if _, err := u.db.ExecContext(ctx, query, user.ID, user.Email, user.PasswordHash, user.CreatedAt); err != nil {
		return fmt.Errorf("couldnt create user: %w", err)
	}
	return nil
}

func (u *UserRepository) GetUserById(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id,email,created_at FROM users WHERE is=$1`
	err := u.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("couldnt get user: %w", err)
	}
	return user, nil

}

func (u *UserRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id=$1`
	u.db.ExecContext(ctx, query, id)
	if _, err := u.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("user could not be deleted: %w", err)
	}
	return nil
}
