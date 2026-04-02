package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Tx(
	ctx context.Context,
	pool *pgxpool.Pool,
	level pgx.TxIsoLevel,
	fns ...func(tx pgx.Tx) error,
) (err error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: level,
	})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	for _, fn := range fns {
		if fn == nil {
			err = errors.New("transaction function is nil")
			return
		}
		if e := fn(tx); e != nil {
			err = e
			return
		}
	}

	return nil
}
