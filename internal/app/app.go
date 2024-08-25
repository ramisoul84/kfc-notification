package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/ramisoul84/kfc-notification/internal/config"
	"github.com/ramisoul84/kfc-notification/internal/notifier"
	emailprovider "github.com/ramisoul84/kfc-notification/internal/provider/email_provider"
	grpcTransport "github.com/ramisoul84/kfc-notification/internal/transport/grpc"
	"github.com/ramisoul84/kfc-notification/pkg/logger"
)

type App struct {
	config     *config.Config
	logger     *logger.Logger
	grpcServer *grpcTransport.Transport
}

func New(cfg *config.Config) (*App, error) {
	log := logger.New(&logger.Config{
		Level:   cfg.Logger.Level,
		Format:  cfg.Logger.Format,
		Output:  cfg.Logger.Output,
		Service: cfg.Logger.Service,
	})

	log.Info("starting kfc-notification",
		"version", cfg.App.Version,
		"environment", cfg.App.Environment,
	)

	// Provider (plumbing)
	emailProvider, err := emailprovider.NewEmailProvider(cfg.Email, log)
	if err != nil {
		return nil, fmt.Errorf("email provider: %w", err)
	}
	log.Info("email provider ready", "provider", cfg.Email.Provider)

	// Notifier (business)
	n, err := notifier.New(emailProvider, log)
	if err != nil {
		return nil, fmt.Errorf("notifier: %w", err)
	}

	// gRPC transport
	notificationServer := grpcTransport.NewNotificationServer(n, log)
	grpcSrv := grpcTransport.NewTransport(cfg, log, notificationServer)

	return &App{
		config:     cfg,
		logger:     log,
		grpcServer: grpcSrv,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := a.grpcServer.Start(); err != nil {
			errCh <- fmt.Errorf("grpc: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
		return nil
	case err := <-errCh:
		a.logger.Error("transport failed", "error", err)
		return err
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down kfc-notification")

	if err := a.grpcServer.Shutdown(ctx); err != nil {
		a.logger.Error("grpc shutdown failed", "error", err)
		return errors.Join(fmt.Errorf("grpc shutdown: %w", err))
	}

	a.logger.Info("shutdown complete")
	return nil
}
