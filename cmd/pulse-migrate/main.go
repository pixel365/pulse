package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/pixel365/pulse/internal/db/postgres"
)

var (
	dsn           string
	migrationsDir string
)

func rootCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate database",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			migrationsDir = os.Getenv("POSTGRES_MIGRATIONS_DIR")
			if migrationsDir == "" {
				return fmt.Errorf("POSTGRES_MIGRATIONS_DIR not set")
			}

			pgConfig := postgres.NewConfigFromEnv()
			if err := pgConfig.Validate(); err != nil {
				return err
			}

			dsn = pgConfig.DSN()

			return nil
		},
	}

	cmd.AddCommand(
		newUpCommand(ctx, &dsn, &migrationsDir),
		newUpToCommand(ctx, &dsn, &migrationsDir),
		newDownCommand(ctx, &dsn, &migrationsDir),
		newDownToCommand(ctx, &dsn, &migrationsDir),
	)

	return cmd
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = godotenv.Load()

	if err := rootCommand(ctx).Execute(); err != nil {
		stop()
		os.Exit(1)
	}
}
