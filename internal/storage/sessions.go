package storage

import (
	"context"
	"database/sql"

	"dota2-hero-grid-maker-back/internal/domain"
	"github.com/google/uuid"
)

type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

func (s *SessionStore) Create(ctx context.Context, userID uuid.UUID) (string, error) {
	const query = `
		INSERT INTO sessions (id, user_id)
		VALUES ($1, $2)
	`

	token := uuid.New()
	if _, err := s.db.ExecContext(ctx, query, token, userID); err != nil {
		return "", err
	}

	return token.String(), nil
}

func (s *SessionStore) GetUserID(ctx context.Context, token uuid.UUID) (uuid.UUID, error) {
	const query = `
		SELECT user_id
		FROM sessions
		WHERE id = $1
	`

	var userID uuid.UUID
	if err := s.db.QueryRowContext(ctx, query, token).Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return uuid.UUID{}, domain.ErrSessionNotFound
		}
		return uuid.UUID{}, err
	}

	return userID, nil
}
