package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	checkrepo "github.com/pixel365/pulse/internal/repository/check"
	checksvc "github.com/pixel365/pulse/internal/services/check"

	"github.com/pixel365/pulse/internal/app"

	"github.com/pixel365/pulse/internal/config"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.MustLoad()

	repo := checkrepo.NewStateRepository()
	stateSvc := checksvc.NewStateService(repo)
	checkHandlerSvc := checksvc.NewHandlerService(stateSvc)

	runner := app.NewApp(cfg, checkHandlerSvc)
	if err := runner.Run(ctx); err != nil {
		stop()
		log.Fatalf("app run error: %v", err)
	}
}
