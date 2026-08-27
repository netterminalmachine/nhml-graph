package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	os.Exit(Main())
}

// Main is the CLI entrypoint, extracted so tests can invoke it without os.Exit.
func Main() int {
	app := &cli.App{
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
				Action:  ApplyMigrations,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
		return 1
	}

	return 0
}
