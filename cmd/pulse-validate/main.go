package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/pixel365/pulse/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = godotenv.Load()

	cmd := &cobra.Command{
		Use:   "pulse-validate",
		Short: "Validate configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			_, err := config.Load()
			return err
		},
	}

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}

	println("ok")
}
