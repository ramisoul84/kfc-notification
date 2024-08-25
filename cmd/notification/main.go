package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ramisoul84/kfc-notification/internal/app"
	"github.com/ramisoul84/kfc-notification/internal/config"
)

func main() {
	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Build the application.
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}

	// Cancel ctx on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run until a signal arrives or a transport fails.
	exitCode := 0
	if err := application.Start(ctx); err != nil {
		log.Printf("application error: %v", err)
		exitCode = 1
	}

	// Shutdown with its own timeout, independent of the signal ctx.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.GRPC.ShutdownTimeout,
	)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
		exitCode = 1
	}

	os.Exit(exitCode)
}
