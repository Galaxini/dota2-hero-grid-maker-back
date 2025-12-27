package httpapi

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"dota2-hero-grid-maker-back/internal/domain"
	"github.com/google/uuid"
)

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.Sessions == nil {
			http.Error(w, "auth not configured", http.StatusInternalServerError)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing authorization", http.StatusUnauthorized)
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, prefix))
		token, err := uuid.Parse(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		userID, err := h.Sessions.GetUserID(r.Context(), token)
		if err != nil {
			if errors.Is(err, domain.ErrSessionNotFound) {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			log.Printf("auth middleware error: %v", err)
			http.Error(w, "failed to authenticate", http.StatusInternalServerError)
			return
		}

		ctx := WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
