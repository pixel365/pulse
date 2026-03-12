package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/pixel365/pulse/internal/config"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	_, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = config.MustLoad()
}
