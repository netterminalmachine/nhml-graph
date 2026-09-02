package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// newFakeModule builds a temp tree that looks like this repo:
//
//	<temp>/                 (module root: has go.mod)
//	  go.mod
//	  internal/core/        (typical go-test CWD)
func newFakeModule(t *testing.T) (moduleRoot, packageDir string) {
	t.Helper()

	moduleRoot = t.TempDir()
	packageDir = filepath.Join(moduleRoot, "internal", "core")
	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(moduleRoot, "go.mod"),
		[]byte("module example\n"),
		0644,
	))
	return moduleRoot, packageDir
}

func writeEnv(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, ".env")
	require.NoError(t, os.WriteFile(path, []byte("PG_PORT=5432\n"), 0644))
	return path
}

func Test_findEnvFile_fromPackageDir_findsEnvAtModuleRoot(t *testing.T) {
	// Layout:
	//   moduleRoot/
	//     go.mod
	//     .env                 <-- only .env; this is what we must find
	//     internal/core/       <-- CWD (as when running go test ./internal/core)
	moduleRoot, packageDir := newFakeModule(t)
	envAtModuleRoot := writeEnv(t, moduleRoot)

	t.Chdir(packageDir)

	require.Equal(t, envAtModuleRoot, findEnvFile())
}

func Test_findEnvFile_fromPackageDir_noEnvInModule_returnsEmpty(t *testing.T) {
	// Layout:
	//   moduleRoot/
	//     go.mod               <-- search stops here, even if a parent has .env
	//     internal/core/       <-- CWD
	_, packageDir := newFakeModule(t)

	t.Chdir(packageDir)

	require.Empty(t, findEnvFile())
}

func Test_isPostgresURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "valid postgres url", url: "postgres://user:pass@host:5432/db", want: true},
		{name: "valid postgresql url", url: "postgresql://user:pass@host:5432/db", want: true},
		{name: "invalid url", url: "postgres://", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isPostgresURL(tt.url))
		})
	}
}
