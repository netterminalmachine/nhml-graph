package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rogpeppe/go-internal/testscript"
)

// testscripts share one Postgres database and run in parallel (testscript
// always calls t.Parallel). Serialize DB use and reset schema between scripts.
var testDBMu sync.Mutex

// build an executable just for testing. Named 'nhmlg:test' to minimize any possible collission with aliases in user space.
func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"nhmlg:test": Main,
	}))
}

func TestCommands(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testdata",
		Setup: passDBEnv,
	})
}

// passDBEnv copies Postgres settings from the host process into each script's
// environment. testscript starts with a clean env (it does not inherit PG_*),
// so without this, nhmlg:test could not reach the database that make/CI already
// configured.
//
// The CLI reads PG_DB (not PG_DB_TEST). We deliberately set PG_DB to same value as
// PG_DB_TEST so scripts never touch the primary database.
//
// MIGPATH is intentionally not forwarded: each script should set it to a
// directory inside $WORK (e.g. env MIGPATH=migrations) so tests never touch
// the real migrations folder.
//
// Before each script runs, the test database public schema is dropped and
// recreated so prior scripts cannot leak tables or ledger rows.
func passDBEnv(e *testscript.Env) error {
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	port := os.Getenv("PG_PORT")
	testDB := os.Getenv("PG_DB_TEST")

	for _, kv := range []struct{ key, val string }{
		{"PG_USER", user},
		{"PG_PASSWORD", password},
		{"PG_PORT", port},
	} {
		if kv.val != "" {
			e.Vars = append(e.Vars, kv.key+"="+kv.val)
		}
	}

	if testDB == "" {
		return fmt.Errorf("PG_DB_TEST must be set so testscripts use the test database")
	}
	e.Vars = append(e.Vars, "PG_DB="+testDB, "PG_DB_TEST="+testDB)

	testDBMu.Lock()
	e.Defer(testDBMu.Unlock)

	if err := resetTestDB(user, password, port, testDB); err != nil {
		return fmt.Errorf("reset test database %q: %w", testDB, err)
	}

	return nil
}

func resetTestDB(user, password, port, db string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", user, password, port, db)
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)

	// Wipe all objects from prior scripts; recreate a usable public schema.
	_, err = conn.Exec(ctx, `
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO CURRENT_USER;
		GRANT ALL ON SCHEMA public TO public;
	`)
	if err != nil {
		return fmt.Errorf("drop/create public schema: %w", err)
	}

	return nil
}
