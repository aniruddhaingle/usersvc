package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"usersvc/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrUserNotFound = errors.New("user not found")
var ErrDuplicateEmail = errors.New("email already exists")

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id,email,password_hash,created_at) VALUES ($1,$2,$3,$4)`
	if _, err := u.db.ExecContext(ctx, query, user.ID, user.Email, user.PasswordHash, user.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("couldnt create user: %w", ErrDuplicateEmail)
		}
		return fmt.Errorf("couldnt create user: %w", err)
	}
	return nil
}

func (u *UserRepository) GetUserById(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id,email,created_at FROM users WHERE id=$1`
	err := u.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("couldnt get user: %w", err)
	}
	return user, nil

}

func (u *UserRepository) ListUsers(ctx context.Context) ([]*domain.User, error) {
	query := `SELECT id,email,created_at FROM users ORDER BY created_at DESC`
	rows, err := u.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("couldnt list users: %w", err)
	}
	defer rows.Close()
	users := []*domain.User{}
	for rows.Next() {
		user := &domain.User{}
		if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("couldnt scan user row: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (u *UserRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id=$1`
	res, err := u.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("user could not be deleted: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("couldnt check deleted rows: %w", err)
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}
