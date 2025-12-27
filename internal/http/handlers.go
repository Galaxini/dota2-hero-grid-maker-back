package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"dota2-hero-grid-maker-back/internal/auth"
	"dota2-hero-grid-maker-back/internal/domain"
	"dota2-hero-grid-maker-back/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Handler struct {
	Grids    *storage.GridRepository
	Sessions *storage.SessionStore
	Users    *storage.UserStore
}

type contextKey string

const userIDKey contextKey = "userID"

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if h.Users == nil {
		http.Error(w, "storage not configured", http.StatusInternalServerError)
		return
	}

	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("hash password error: %v", err)
		http.Error(w, "failed to register", http.StatusInternalServerError)
		return
	}

	if _, err := h.Users.Create(r.Context(), req.Email, hash); err != nil {
		if isUniqueViolation(err) {
			http.Error(w, "email already registered", http.StatusConflict)
			return
		}
		log.Printf("create user error: %v", err)
		http.Error(w, "failed to register", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if h.Users == nil || h.Sessions == nil {
		http.Error(w, "storage not configured", http.StatusInternalServerError)
		return
	}

	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	token, err := auth.Login(r.Context(), h.Users, h.Sessions, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		log.Printf("login error: %v", err)
		http.Error(w, "failed to login", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

func (h *Handler) handleGetDefaultGrid(w http.ResponseWriter, r *http.Request) {
	if h.Grids == nil {
		http.Error(w, "storage not configured", http.StatusInternalServerError)
		return
	}

	grid, err := h.Grids.GetDefault(r.Context())
	if err != nil {
		if errors.Is(err, domain.ErrGridNotFound) || errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "default grid not found", http.StatusNotFound)
			return
		}
		log.Printf("get default grid error: %v", err)
		http.Error(w, "failed to fetch default grid", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, gridResponseFromDomain(grid))
}

func (h *Handler) handleCreateGrid(w http.ResponseWriter, r *http.Request) {
	if h.Grids == nil {
		http.Error(w, "storage not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createGridRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Title == "" || len(req.Data) == 0 {
		http.Error(w, "title and data are required", http.StatusBadRequest)
		return
	}

	grid, err := h.Grids.Create(r.Context(), userID, req.Title, req.Data)
	if err != nil {
		log.Printf("create grid error: %v", err)
		http.Error(w, "failed to create grid", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, gridResponseFromDomain(grid))
}

func (h *Handler) handleListGrids(w http.ResponseWriter, r *http.Request) {
	if h.Grids == nil {
		http.Error(w, "storage not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	grids, err := h.Grids.GetByUser(r.Context(), userID)
	if err != nil {
		log.Printf("list grids error: %v", err)
		http.Error(w, "failed to fetch grids", http.StatusInternalServerError)
		return
	}

	resp := make([]gridResponse, 0, len(grids))
	for _, g := range grids {
		resp = append(resp, gridResponseFromDomain(g))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleGetGrid(w http.ResponseWriter, r *http.Request) {
	if h.Grids == nil {
		http.Error(w, "storage not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	gridIDParam := chi.URLParam(r, "id")
	gridID, err := uuid.Parse(gridIDParam)
	if err != nil {
		http.Error(w, "invalid grid id", http.StatusBadRequest)
		return
	}

	grid, err := h.Grids.GetByID(r.Context(), userID, gridID)
	if err != nil {
		if errors.Is(err, domain.ErrGridNotFound) || errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "grid not found", http.StatusNotFound)
			return
		}
		log.Printf("get grid error: %v", err)
		http.Error(w, "failed to fetch grid", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, gridResponseFromDomain(grid))
}

type createGridRequest struct {
	Title string          `json:"title"`
	Data  json.RawMessage `json:"data"`
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type gridResponse struct {
	ID        uuid.UUID       `json:"id"`
	UserID    *uuid.UUID      `json:"user_id,omitempty"`
	Title     string          `json:"title"`
	Data      json.RawMessage `json:"data"`
	CreatedAt string          `json:"created_at"`
}

func gridResponseFromDomain(g *domain.Grid) gridResponse {
	return gridResponse{
		ID:        g.ID,
		UserID:    g.UserID,
		Title:     g.Title,
		Data:      g.Data,
		CreatedAt: g.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(v); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("invalid JSON payload")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}
