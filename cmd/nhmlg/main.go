package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	os.Exit(Main())
}

// Main is the CLI entrypoint, extracted so tests can invoke it without os.Exit.
func Main() int {
	// set up logging
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug, // Set minimum log level
		AddSource: true,            // Includes file name and line number
	}
	l := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(l)

	cmd := &cli.Command{
		Name:  "nhmlg",
		Usage: "nhml-graph (nhmlg) is a tiny, env-aware, forward-only, postgres migrations manager.",
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialise the migrations table",
				Action:  InitialiseMigrations,
			},
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create a new migration",
				Action:  CreateNewMigration,
			},
			{
				Name:    "migrate",
				Aliases: []string{"m"},
				Usage:   "Apply all pending (un-applied) migrations",
				Action:  MigrateAll,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("error running app", "error", err)
		return 1
	}

	return 0
}
