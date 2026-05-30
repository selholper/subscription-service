package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"subscription-service/internal/app"
	"subscription-service/internal/config"
	"subscription-service/internal/logger"

	_ "subscription-service/docs"
)

// @title           Subscription Service API
// @version         1.0
// @description     REST service for aggregating data about users' online subscriptions.
// @host            localhost:8080
// @BasePath        /
func main() {
	envFile := flag.String("env", ".env", "path to .env file")
	flag.Parse()

	cfg, err := config.Load(*envFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	application, err := app.New(cfg, log)
	if err != nil {
		log.Fatal("failed to initialize application", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err = application.Run(ctx); err != nil {
		log.Fatal("application run failed", zap.Error(err))
	}
}
