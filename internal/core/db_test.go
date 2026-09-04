package core

import (
	"os"
	"testing"

	"github.com/netterminalmachine/nhml-graph/internal/testdb"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Exit(testdb.RunTests(m))
}

func TestRunSingleMigration(t *testing.T) {
	ctx := t.Context()
	pool := testdb.Pool(t)
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer func() {
		err := tx.Rollback(ctx)
		require.NoError(t, err)
	}()

	err = runSingleMigration(ctx, tx, []string{
		"CREATE TABLE IF NOT EXISTS run_single_migration (id SERIAL PRIMARY KEY, name TEXT);",
		"INSERT INTO run_single_migration (name) VALUES ('test');",
	})
	require.NoError(t, err)

	rows, err := tx.Query(ctx, "SELECT name FROM run_single_migration")
	require.NoError(t, err)
	defer rows.Close()

	var name string
	for rows.Next() {
		err := rows.Scan(&name)
		require.NoError(t, err)
		require.Equal(t, "test", name)
	}
}
