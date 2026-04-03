package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type TxManager interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

func Tx(
	ctx context.Context,
	manager TxManager,
	level pgx.TxIsoLevel,
	fns ...func(tx pgx.Tx) error,
) (err error) {
	tx, err := manager.BeginTx(ctx, pgx.TxOptions{
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
