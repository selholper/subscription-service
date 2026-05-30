package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"subscription-service/internal/config"
	"subscription-service/internal/database"
	"subscription-service/internal/repository/postgres"
	"subscription-service/internal/service"
	transport "subscription-service/internal/transport/http"
)

type App struct {
	cfg    *config.Config
	logger *zap.Logger
	server *http.Server
}

func New(cfg *config.Config, logger *zap.Logger) (*App, error) {
	db, err := database.NewPostgres(database.Options{
		DSN:             cfg.DSN,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		return nil, err
	}

	repo := postgres.NewSubscriptionRepository(db)
	svc := service.NewSubscriptionService(repo, logger)
	handler := transport.NewSubscriptionHandler(svc, logger)
	router := transport.NewRouter(handler, logger)

	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		cfg:    cfg,
		logger: logger,
		server: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		a.logger.Info("http server started", zap.String("addr", a.server.Addr))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
		return a.shutdown()
	case err := <-errCh:
		return err
	}
}

func (a *App) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("graceful shutdown failed", zap.Error(err))
		return err
	}

	a.logger.Info("server stopped gracefully")
	return nil
}
