package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/waynenilsen/waynebot/internal/agent"
	"github.com/waynenilsen/waynebot/internal/api"
	"github.com/waynenilsen/waynebot/internal/config"
	"github.com/waynenilsen/waynebot/internal/connector"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/llm"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/tools"
	"github.com/waynenilsen/waynebot/internal/ws"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	v, err := database.SchemaVersion()
	if err != nil {
		slog.Error("failed to get schema version", "error", err)
		os.Exit(1)
	}
	slog.Info("database ready", "schema_version", v)

	hub := ws.NewHub()
	go hub.Run()

	llmClient := llm.NewClient(cfg.OpenRouterKey)
	toolsRegistry := tools.NewRegistry()
	toolsRegistry.RegisterDefaults(&tools.SandboxConfig{
		BaseDir:      cfg.SandboxDir,
		BlockedHosts: tools.DefaultBlockedHosts(),
		AllowedCommands: []string{
			"ls", "cat", "head", "tail", "grep", "find", "wc",
			"sort", "uniq", "diff", "echo", "date", "pwd",
		},
	})
	supervisor := agent.NewSupervisor(database, hub, llmClient, toolsRegistry)

	if err := supervisor.StartAll(); err != nil {
		slog.Error("failed to start agent supervisor", "error", err)
		os.Exit(1)
	}
	slog.Info("agent supervisor started")

	connectors := connector.NewRegistry()
	if cfg.IMAPHost != "" {
		ch, err := model.GetChannelByName(database, cfg.IMAPChannel)
		if err != nil {
			ch, err = model.CreateChannel(database, cfg.IMAPChannel, "incoming email")
			if err != nil {
				slog.Error("failed to create email channel", "error", err)
				os.Exit(1)
			}
		}
		emailConn := connector.NewEmailConnector(connector.EmailConfig{
			Host:      cfg.IMAPHost,
			Port:      cfg.IMAPPort,
			User:      cfg.IMAPUser,
			Pass:      cfg.IMAPPass,
			ChannelID: ch.ID,
		}, connector.NewTLSIMAPClient(), database, hub)
		connectors.Register(emailConn)
		slog.Info("email connector registered", "user", cfg.IMAPUser)
	}
	connectors.StartAll()

	router := api.NewRouter(database, cfg.CORSOrigins, hub, supervisor)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Session and ws_ticket cleanup goroutine.
	go runCleanup(ctx, database)

	archiver := &agent.Archiver{DB: database, ArchiveDir: cfg.ArchiveDir}
	go archiver.Run(ctx)
	slog.Info("archiver started", "archive_dir", cfg.ArchiveDir)

	go func() {
		slog.Info("listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	// Shutdown order: HTTP → connectors → agents → WS hub → DB (via defer).
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown error", "error", err)
	}
	slog.Info("http server stopped")

	connectors.StopAll()
	slog.Info("connectors stopped")

	supervisor.StopAll()
	slog.Info("agent supervisor stopped")

	hub.Stop()
	slog.Info("ws hub stopped")

	slog.Info("stopped")
}

// runCleanup deletes expired sessions and ws_tickets every 15 minutes.
func runCleanup(ctx context.Context, database *db.DB) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sessions, err := model.CleanupExpiredSessions(database)
			if err != nil {
				slog.Error("session cleanup failed", "error", err)
			} else if sessions > 0 {
				slog.Info("cleaned expired sessions", "count", sessions)
			}

			tickets, err := model.CleanupExpiredWsTickets(database)
			if err != nil {
				slog.Error("ws_ticket cleanup failed", "error", err)
			} else if tickets > 0 {
				slog.Info("cleaned expired ws_tickets", "count", tickets)
			}
		}
	}
}
