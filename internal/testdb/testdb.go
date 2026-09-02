// Package testdb provides a shared Postgres test database for integration tests.
//
// testdb must not import internal/core: core tests import this helper, and a
// reverse import would be a cycle.
//
// The test database must already exist (created via devtools/init.sql on first
// docker compose up). Wire TestMain in any package that runs DB integration tests:
//
//	func TestMain(m *testing.M) {
//		os.Exit(testdb.RunTests(m))
//	}
//
// Individual tests then use the shared pool:
//
//	func TestSomething(t *testing.T) {
//		pool := testdb.Pool(t)
//		// ...
//	}
//
// Schema migrations are optional. Assign RunMigrations before TestMain if the
// tests need the migrated schema (from core tests, no extra import is needed):
//
//	testdb.RunMigrations = RunMigrations
package testdb

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/netterminalmachine/nhml-graph/internal/helpers"
)

var testPool *pgxpool.Pool

// RunMigrations, if set, is called during setup after the pool is created.
var RunMigrations func(ctx context.Context, config *helpers.Config, pool *pgxpool.Pool) error

// RunTests connects to the test database, runs tests, then closes the pool.
// Call this from TestMain in any package that uses Pool.
func RunTests(m *testing.M) int {
	if err := setup(); err != nil {
		fmt.Fprintf(os.Stderr, "testdb: setup failed: %v\n", err)
		return 1
	}

	code := m.Run()
	teardown()

	return code
}

// Pool returns the shared test database pool.
// TestMain must call RunTests before any test uses Pool.
func Pool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if testPool == nil {
		t.Fatal("testdb: Pool called before setup; add TestMain that calls testdb.RunTests")
	}

	return testPool
}

func setup() error {
	config, err := helpers.LoadConfig(helpers.WithTestDB())
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, config.PgUrl)
	if err != nil {
		return fmt.Errorf("create test db connection pool: %w", err)
	}

	if RunMigrations != nil {
		err = RunMigrations(ctx, config, pool)
		if err != nil {
			pool.Close()
			return fmt.Errorf("run migrations on test db failed: %w", err)
		}
	}

	testPool = pool
	return nil
}

func teardown() {
	if testPool != nil {
		testPool.Close()
		testPool = nil
	}
}
