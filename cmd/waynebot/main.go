package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/waynenilsen/waynebot/internal/agent"
	"github.com/waynenilsen/waynebot/internal/api"
	"github.com/waynenilsen/waynebot/internal/config"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/llm"
	"github.com/waynenilsen/waynebot/internal/tools"
	"github.com/waynenilsen/waynebot/internal/ws"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	v, err := database.SchemaVersion()
	if err != nil {
		log.Fatalf("failed to get schema version: %v", err)
	}
	log.Printf("database ready, schema version %d", v)

	hub := ws.NewHub()
	go hub.Run()

	llmClient := llm.NewClient(cfg.OpenRouterKey)
	toolsRegistry := tools.NewRegistry()
	supervisor := agent.NewSupervisor(database, hub, llmClient, toolsRegistry)

	if err := supervisor.StartAll(); err != nil {
		log.Fatalf("failed to start agent supervisor: %v", err)
	}
	log.Println("agent supervisor started")

	router := api.NewRouter(database, cfg.CORSOrigins, hub, supervisor)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("listening on :%d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	supervisor.StopAll()
	log.Println("agent supervisor stopped")

	hub.Stop()

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("stopped")
}
