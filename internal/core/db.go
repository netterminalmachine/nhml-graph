package core

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const MigrationsTableSQL = `
create table if not exists migrations (
	id int primary key,
	name text not null,
	hash text not null
);
`

func runSingleMigration(ctx context.Context, tx pgx.Tx, sqlStrings []string) error {
	for _, sql := range sqlStrings {
		_, err := tx.Exec(ctx, sql)
		if err != nil {
			return fmt.Errorf("SQL execution error: %w", err)
		}
	}

	return nil
}

func asTransactionWithAutoRollback(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = fn(tx)
	if err != nil {
		slog.Error("encountered error in transaction - rolling back", "error", err)
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			slog.Error("💣Uhm... so... the rollback ALSO failed. OMG. 💀", "error", errRollback)
			return errRollback
		}
		slog.Info("transaction rollback OK")
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		slog.Error("❌Dagnammit... the commit transaction failed!! 😫", "error", err)
		return err
	}
	return nil
}
