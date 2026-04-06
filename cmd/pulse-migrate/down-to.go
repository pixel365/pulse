package main

import (
	"context"
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func newDownToCommand(ctx context.Context, dsn, dir *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down-to",
		Short: "Migrate down to a specific version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := goose.OpenDBWithDriver("pgx", *dsn)
			if err != nil {
				return err
			}

			defer func() {
				_ = db.Close()
			}()

			version, err := cmd.Flags().GetInt64("version")
			if err != nil {
				return err
			}

			if version <= 0 {
				return fmt.Errorf("version must be positive")
			}

			if err = goose.SetDialect("pgx"); err != nil {
				return err
			}

			return goose.DownToContext(ctx, db, *dir, version)
		},
	}

	cmd.Flags().Int64P("version", "v", 0, "version to migrate down to")
	_ = cmd.MarkFlagRequired("version")

	return cmd
}
