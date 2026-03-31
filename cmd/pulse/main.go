package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

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

	runner := app.NewApp(cfg)
	if err := runner.Run(ctx); err != nil {
		stop()
		log.Fatalf("app run error: %v", err)
	}
}
