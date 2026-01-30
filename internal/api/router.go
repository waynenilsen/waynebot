package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/waynenilsen/waynebot/internal/auth"
	"github.com/waynenilsen/waynebot/internal/db"
)

// NewRouter creates the main Chi router with all middleware and route groups.
func NewRouter(database *db.DB, corsOrigins []string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(auth.Middleware(database))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	ah := &AuthHandler{DB: database}

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/register", ah.Register)
		r.Post("/auth/login", ah.Login)
		r.With(auth.RequireAuth).Post("/auth/logout", ah.Logout)
		r.With(auth.RequireAuth).Get("/auth/me", ah.Me)
	})

	return r
}
