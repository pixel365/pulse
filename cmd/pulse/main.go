package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/pixel365/pulse/internal/db/postgres"

	"github.com/pixel365/pulse/internal/logger"

	checkrepo "github.com/pixel365/pulse/internal/repository/check"
	checksvc "github.com/pixel365/pulse/internal/services/check"

	"github.com/pixel365/pulse/internal/app"

	"github.com/pixel365/pulse/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = godotenv.Load()

	cfg := config.MustLoad()
	log := logger.NewSlog()

	pgConfig := postgres.NewConfigFromEnv()
	pgPool, err := postgres.NewPool(ctx, pgConfig)
	if err != nil {
		log.Error(ctx, "postgres pool error", "error", err)
		return
	}
	defer pgPool.Close()

	staterepo := checkrepo.NewStateRepository(pgPool)
	execRepo := checkrepo.NewExecutionRepository(pgPool)
	stateSvc := checksvc.NewStateService(staterepo)
	checkHandlerSvc := checksvc.NewHandlerService(stateSvc, execRepo)

	runner := app.NewApp(cfg, log, checkHandlerSvc)
	if err := runner.Run(ctx); err != nil {
		stop()
		log.Error(ctx, "failed to run app", "error", err)
		os.Exit(1)
	}
}
