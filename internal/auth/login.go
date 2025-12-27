package auth

import (
	"context"
	"database/sql"
	"errors"

	"dota2-hero-grid-maker-back/internal/storage"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

func Login(ctx context.Context, users *storage.UserStore, sessions *storage.SessionStore, email, password string) (string, error) {
	user, err := users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := CheckPassword(user.PasswordHash, password); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := sessions.Create(ctx, user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}
