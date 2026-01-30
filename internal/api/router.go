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
	ch := &ChannelHandler{DB: database}
	ph := &PersonaHandler{DB: database}
	ih := &InviteHandler{DB: database}

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/register", ah.Register)
		r.Post("/auth/login", ah.Login)
		r.With(auth.RequireAuth).Post("/auth/logout", ah.Logout)
		r.With(auth.RequireAuth).Get("/auth/me", ah.Me)

		r.With(auth.RequireAuth).Get("/channels", ch.ListChannels)
		r.With(auth.RequireAuth).Post("/channels", ch.CreateChannel)
		r.With(auth.RequireAuth).Get("/channels/{id}/messages", ch.GetMessages)
		r.With(auth.RequireAuth).Post("/channels/{id}/messages", ch.PostMessage)

		r.With(auth.RequireAuth).Get("/personas", ph.ListPersonas)
		r.With(auth.RequireAuth).Post("/personas", ph.CreatePersona)
		r.With(auth.RequireAuth).Put("/personas/{id}", ph.UpdatePersona)
		r.With(auth.RequireAuth).Delete("/personas/{id}", ph.DeletePersona)

		r.With(auth.RequireAuth).Post("/invites", ih.CreateInvite)
		r.With(auth.RequireAuth).Get("/invites", ih.ListInvites)
	})

	return r
}
