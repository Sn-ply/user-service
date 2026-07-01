package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/snaply/user-service/internal/service"
	"go.uber.org/zap"
)

func NewRouter(auth service.AuthService, users service.UserService, log *zap.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	authH := NewAuthHandler(auth, log)
	userH := NewUserHandler(users, log)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)
			r.Post("/refresh", authH.Refresh)
			r.Post("/logout", authH.Logout)
		})

		r.Route("/users", func(r chi.Router) {
			r.Get("/search", userH.Search)
			r.Put("/me", userH.UpdateMe)
			r.Post("/batch", userH.Batch)
			r.Get("/{username}", userH.GetProfile)
		})
	})

	return r
}
