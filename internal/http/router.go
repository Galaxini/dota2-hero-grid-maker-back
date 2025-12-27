package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Post("/auth/register", h.handleRegister)
	r.Post("/auth/login", h.handleLogin)

	r.Get("/default-grid", h.handleGetDefaultGrid)

	r.Route("/grids", func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Post("/", h.handleCreateGrid)
		r.Get("/", h.handleListGrids)
		r.Get("/{id}", h.handleGetGrid)
	})

	return r
}
