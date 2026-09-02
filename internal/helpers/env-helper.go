package helpers

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	PgUrl         string
	MigrationsDir string
	MainSchema    string
}

type EnvOptions struct {
	UseTestDb bool
}

type Option func(o *EnvOptions)

var defaultEnvOptions = EnvOptions{
	UseTestDb: false,
}

func WithTestDB() Option {
	return func(o *EnvOptions) {
		o.UseTestDb = true
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		slog.Error("No value for env key", slog.String("key", key))
		return ""
	}
	return value
}

func LoadConfig(opts ...Option) (*Config, error) {
	options := defaultEnvOptions
	for _, opt := range opts {
		opt(&options)
	}

	envpath := findEnvFile()
	if !IsBlank(envpath) {
		err := godotenv.Load(envpath)
		if err != nil {
			slog.Warn("Could not load env file at given path", slog.String("path", envpath))
			return nil, err
		}
	}

	targetDb := getEnv("PG_DB")
	if options.UseTestDb {
		targetDb = getEnv("PG_DB_TEST")
	}

	pgUrl := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", getEnv("PG_USER"), getEnv("PG_PASSWORD"), getEnv("PG_PORT"), targetDb)
	migDir := getEnv("MIGPATH")

	// check for valid connection string
	if !isPostgresURL(pgUrl) {
		return nil, fmt.Errorf("invalid connection string: %s", pgUrl)
	}

	config := Config{
		PgUrl:         pgUrl,
		MigrationsDir: migDir,
	}

	return &config, nil
}

// findEnvFile looks for .env in the current directory and each parent, stopping
// at the Go module root (directory containing go.mod). Tests run with CWD set
// to the package dir, so "./.env" would miss the file at the repo root.
func findEnvFile() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}

		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return ""
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func isPostgresURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}

	return (u.Scheme == "postgres" || u.Scheme == "postgresql") &&
		u.Host != "" &&
		u.Path != ""
}
