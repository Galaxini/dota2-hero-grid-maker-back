package storage

import (
	"context"
	"database/sql"

	"dota2-hero-grid-maker-back/internal/domain"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at
	`

	var u domain.User
	if err := s.db.QueryRowContext(ctx, query, email, passwordHash).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	var u domain.User
	if err := s.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil, err
	}

	return &u, nil
}
