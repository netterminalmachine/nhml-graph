package main

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/urfave/cli/v3"

	"github.com/netterminalmachine/nhml-graph/internal/core"
	"github.com/netterminalmachine/nhml-graph/internal/helpers"
)

func InitialiseMigrations(ctx context.Context, _ *cli.Command) error {
	config, err := helpers.LoadConfig()
	if err != nil {
		return err
	}
	sqlStr := core.MigrationsTableSQL

	conn, err := pgx.Connect(ctx, config.PgUrl)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := conn.Close(ctx)
		if closeErr != nil {
			slog.Error("db connection close error", "error", closeErr)
		}
	}()

	statusText, err := conn.Exec(ctx, sqlStr)
	if err != nil {
		slog.Error("SQL execution error", "error", err, "status text", statusText)
		return err
	}

	slog.Info("Migrations table ready.")

	return nil
}
