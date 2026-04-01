package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

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
	repo := checkrepo.NewStateRepository()
	stateSvc := checksvc.NewStateService(repo)
	checkHandlerSvc := checksvc.NewHandlerService(stateSvc)

	runner := app.NewApp(cfg, log, checkHandlerSvc)
	if err := runner.Run(ctx); err != nil {
		stop()
		log.Error(ctx, "failed to run app", "error", err)
		os.Exit(1)
	}
}
