package main

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v3"

	"github.com/netterminalmachine/nhml-graph/internal/core"
	"github.com/netterminalmachine/nhml-graph/internal/helpers"
)

func MigrateAll(ctx context.Context, _ *cli.Command) error {
	config, err := helpers.LoadConfig()
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, config.PgUrl)
	if err != nil {
		slog.Error("Unable to create connection pool", "error", err)
		return err
	}
	defer pool.Close()

	err = core.RunMigrations(ctx, config, pool)
	if err != nil {
		slog.Error("Unable to run migrations", "error", err)
		return err
	}

	slog.Info("Migrations applied successfully")
	return nil
}
