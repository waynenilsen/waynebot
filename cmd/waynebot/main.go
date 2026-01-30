package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/waynenilsen/waynebot/internal/api"
	"github.com/waynenilsen/waynebot/internal/config"
	"github.com/waynenilsen/waynebot/internal/db"
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

	router := api.NewRouter(database, cfg.CORSOrigins)

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

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("stopped")
}
