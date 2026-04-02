package main

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func newDownCommand(ctx context.Context, dsn, dir *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Migrate down",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := goose.OpenDBWithDriver("pgx", *dsn)
			if err != nil {
				return err
			}

			defer func() {
				_ = db.Close()
			}()

			if err = goose.SetDialect("pgx"); err != nil {
				return err
			}

			return goose.DownContext(ctx, db, *dir)
		},
	}

	return cmd
}
