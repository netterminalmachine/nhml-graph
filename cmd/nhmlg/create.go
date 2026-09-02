package main

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v3"

	"github.com/netterminalmachine/nhml-graph/internal/core"
	"github.com/netterminalmachine/nhml-graph/internal/helpers"
)

func CreateNewMigration(ctx context.Context, cmd *cli.Command) error {
	config, err := helpers.LoadConfig()
	if err != nil {
		return err
	}

	if cmd.Args().Len() < 3 {
		return errors.New("you must provide a name for your migration, which must be at least three words. e.g. 'add foo table'")
	}

	nameArgs := strings.Join(cmd.Args().Slice(), "-")

	pool, err := pgxpool.New(ctx, config.PgUrl)
	if err != nil {
		slog.Error("Unable to create connection pool", "error", err)
		return err
	}
	defer pool.Close()

	_, err = core.CreateMigration(ctx, config, pool, nameArgs)
	if err != nil {
		slog.Error("Unable to create migration", "error", err)
		return err
	}

	slog.Info("Migration created successfully", "name", nameArgs)
	return nil
}
