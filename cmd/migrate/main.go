package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	"subscription-service/internal/config"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

const migrationsDir = "migrations"

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	envFile := flag.String("env", ".env", "path to .env file")
	command := flag.String("cmd", "up", "goose command: up, down, status, reset, version")
	flag.Parse()

	logger.Info("loading config", zap.String("env_file", *envFile))

	cfg, err := config.Load(*envFile)
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	logger.Info("connecting to database")

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		logger.Fatal("failed to open db connection", zap.Error(err))
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}

	logger.Info("database connection established")

	if err = goose.SetDialect("postgres"); err != nil {
		logger.Fatal("failed to set goose dialect", zap.Error(err))
	}

	goose.SetLogger(newGooseLogger(logger))

	args := flag.Args()

	logger.Info("running migration", zap.String("command", *command))

	if err = goose.RunWithOptionsContext(
		context.Background(),
		*command,
		db,
		migrationsDir,
		args,
		goose.WithAllowMissing(),
	); err != nil {
		logger.Fatal("migration failed",
			zap.String("command", *command),
			zap.Error(err),
		)
	}

	logger.Info("migration completed successfully", zap.String("command", *command))
}

type gooseLogger struct {
	log *zap.SugaredLogger
}

func newGooseLogger(logger *zap.Logger) *gooseLogger {
	return &gooseLogger{log: logger.Named("goose").Sugar()}
}

func (g *gooseLogger) Fatalf(format string, v ...any) {
	g.log.Fatalf(format, v...)
}

func (g *gooseLogger) Printf(format string, v ...any) {
	g.log.Infof(format, v...)
}
