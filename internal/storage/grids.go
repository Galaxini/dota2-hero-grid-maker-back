package storage

import (
	"context"
	"database/sql"
	"encoding/json"

	"dota2-hero-grid-maker-back/internal/domain"
	"github.com/google/uuid"
)

type GridRepository struct {
	db *sql.DB
}

func NewGridRepository(db *sql.DB) *GridRepository {
	return &GridRepository{db: db}
}

func (r *GridRepository) GetDefault(ctx context.Context) (*domain.Grid, error) {
	const query = `
		SELECT id, user_id, title, data, created_at
		FROM grids
		WHERE user_id IS NULL
		LIMIT 1
	`

	var g domain.Grid
	var userID uuid.NullUUID
	if err := r.db.QueryRowContext(ctx, query).Scan(&g.ID, &userID, &g.Title, &g.Data, &g.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrGridNotFound
		}
		return nil, err
	}

	if userID.Valid {
		id := userID.UUID
		g.UserID = &id
	}

	return &g, nil
}

func (r *GridRepository) Create(ctx context.Context, userID uuid.UUID, title string, data json.RawMessage) (*domain.Grid, error) {
	const deleteQuery = `
		DELETE FROM grids
		WHERE user_id = $1
	`
	const insertQuery = `
		INSERT INTO grids (user_id, title, data)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, deleteQuery, userID); err != nil {
		return nil, err
	}

	g := domain.Grid{
		Title: title,
		Data:  data,
	}

	if err := tx.QueryRowContext(ctx, insertQuery, userID, title, data).Scan(&g.ID, &g.CreatedAt); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	g.UserID = &userID

	return &g, nil
}

func (r *GridRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Grid, error) {
	const query = `
		SELECT id, user_id, title, data, created_at
		FROM grids
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grids []*domain.Grid
	for rows.Next() {
		var g domain.Grid
		var uid uuid.UUID
		if err := rows.Scan(&g.ID, &uid, &g.Title, &g.Data, &g.CreatedAt); err != nil {
			return nil, err
		}
		gUserID := uid
		g.UserID = &gUserID
		grids = append(grids, &g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return grids, nil
}

func (r *GridRepository) GetByID(ctx context.Context, userID, gridID uuid.UUID) (*domain.Grid, error) {
	const query = `
		SELECT id, user_id, title, data, created_at
		FROM grids
		WHERE id = $1 AND user_id = $2
	`

	var g domain.Grid
	var uid uuid.UUID
	if err := r.db.QueryRowContext(ctx, query, gridID, userID).Scan(&g.ID, &uid, &g.Title, &g.Data, &g.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrGridNotFound
		}
		return nil, err
	}
	gUserID := uid
	g.UserID = &gUserID

	return &g, nil
}
