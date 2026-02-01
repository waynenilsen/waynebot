package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/waynenilsen/waynebot/internal/agent"
	"github.com/waynenilsen/waynebot/internal/auth"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// NewRouter creates the main Chi router with all middleware and route groups.
func NewRouter(database *db.DB, corsOrigins []string, hub *ws.Hub, supervisor ...*agent.Supervisor) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	limiter := NewRateLimiter(60, 120)
	r.Use(limiter.Middleware)
	r.Use(middleware.Logger)
	r.Use(auth.Middleware(database))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	ah := &AuthHandler{DB: database}
	var sup *agent.Supervisor
	if len(supervisor) > 0 {
		sup = supervisor[0]
	}
	ch := &ChannelHandler{DB: database, Hub: hub, Supervisor: sup}
	ph := &PersonaHandler{DB: database}
	ih := &InviteHandler{DB: database}
	wh := &WsHandler{DB: database, Hub: hub}

	dh := &DMHandler{DB: database, Hub: hub, Supervisor: sup}
	mh := &MemberHandler{DB: database, Supervisor: sup}
	uh := &UserHandler{DB: database}

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/register", ah.Register)
		r.Post("/auth/login", ah.Login)
		r.With(auth.RequireAuth).Post("/auth/logout", ah.Logout)
		r.With(auth.RequireAuth).Get("/auth/me", ah.Me)

		r.With(auth.RequireAuth).Get("/channels", ch.ListChannels)
		r.With(auth.RequireAuth).Post("/channels", ch.CreateChannel)
		r.With(auth.RequireAuth).Get("/channels/{id}/messages", ch.GetMessages)
		r.With(auth.RequireAuth).Post("/channels/{id}/messages", ch.PostMessage)
		r.With(auth.RequireAuth).Post("/channels/{id}/read", ch.MarkRead)

		r.With(auth.RequireAuth).Get("/channels/{id}/members", mh.ListMembers)
		r.With(auth.RequireAuth).Post("/channels/{id}/members", mh.AddMember)
		r.With(auth.RequireAuth).Delete("/channels/{id}/members", mh.RemoveMember)

		cph := &ChannelProjectHandler{DB: database}
		r.With(auth.RequireAuth).Get("/channels/{id}/projects", cph.ListChannelProjects)
		r.With(auth.RequireAuth).Post("/channels/{id}/projects", cph.AddChannelProject)
		r.With(auth.RequireAuth).Delete("/channels/{id}/projects/{projectID}", cph.RemoveChannelProject)

		rh := &ReactionHandler{DB: database, Hub: hub}
		r.With(auth.RequireAuth).Put("/channels/{id}/messages/{messageID}/reactions", rh.AddReaction)
		r.With(auth.RequireAuth).Delete("/channels/{id}/messages/{messageID}/reactions", rh.RemoveReaction)

		r.With(auth.RequireAuth).Get("/personas", ph.ListPersonas)
		r.With(auth.RequireAuth).Post("/personas", ph.CreatePersona)
		r.With(auth.RequireAuth).Put("/personas/{id}", ph.UpdatePersona)
		r.With(auth.RequireAuth).Delete("/personas/{id}", ph.DeletePersona)

		prh := &ProjectHandler{DB: database}
		r.With(auth.RequireAuth).Get("/projects", prh.ListProjects)
		r.With(auth.RequireAuth).Post("/projects", prh.CreateProject)
		r.With(auth.RequireAuth).Put("/projects/{id}", prh.UpdateProject)
		r.With(auth.RequireAuth).Delete("/projects/{id}", prh.DeleteProject)

		r.With(auth.RequireAuth).Post("/invites", ih.CreateInvite)
		r.With(auth.RequireAuth).Get("/invites", ih.ListInvites)

		r.With(auth.RequireAuth).Get("/users", uh.ListUsers)

		mtnh := &MentionHandler{DB: database}
		r.With(auth.RequireAuth).Get("/mention-targets", mtnh.ListMentionTargets)

		r.With(auth.RequireAuth).Get("/dms", dh.ListDMs)
		r.With(auth.RequireAuth).Post("/dms", dh.CreateDM)

		r.With(auth.RequireAuth).Post("/ws/ticket", wh.CreateTicket)

		if len(supervisor) > 0 && supervisor[0] != nil {
			agh := &AgentHandler{DB: database, Supervisor: supervisor[0]}
			r.With(auth.RequireAuth).Get("/agents/status", agh.Status)
			r.With(auth.RequireAuth).Post("/agents/start", agh.Start)
			r.With(auth.RequireAuth).Post("/agents/stop", agh.Stop)
			r.With(auth.RequireAuth).Get("/agents/{persona_id}/llm-calls", agh.LLMCalls)
			r.With(auth.RequireAuth).Get("/agents/{persona_id}/tool-executions", agh.ToolExecutions)
			r.With(auth.RequireAuth).Get("/agents/{persona_id}/stats", agh.Stats)

			ctxh := &ContextHandler{DB: database, Hub: hub, Supervisor: supervisor[0]}
			r.With(auth.RequireAuth).Get("/agents/{persona_id}/context-budget", ctxh.ContextBudget)
			r.With(auth.RequireAuth).Post("/agents/{persona_id}/channels/{channel_id}/reset-context", ctxh.ResetContext)
		}
	})

	r.Get("/ws", wh.Upgrade)

	return r
}
