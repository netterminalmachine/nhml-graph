package helpers

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	PgUrl         string
	MigrationsDir string
	MainSchema    string
}

func getEnvWithDefaultAndBlankableFlag(key string, defaultValue string, canBeBlank bool) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		if !canBeBlank && defaultValue == "" {
			msg := fmt.Sprintf("No value for env key %s which cannot be blank.", key)
			log.Fatal(msg)
		}
		return defaultValue
	}
	return value
}

func getEnvWithDefault(key string, defaultValue string) string {
	return getEnvWithDefaultAndBlankableFlag(key, defaultValue, false)
}

func getEnv(key string) string {
	return getEnvWithDefault(key, "")
}

func LoadConfig() *Config {
	// default
	envpath := "./.env"

	if !IsBlank(envpath) {
		err := godotenv.Load(envpath)
		if err != nil {
			msg := fmt.Sprintf("Could not load env file at path [%s]", envpath)
			log.Fatal(msg)
		}
	}

	pgUrl := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", getEnv("PG_USER"), getEnv("PG_PASSWORD"), getEnv("PG_PORT"), getEnv("PG_DB"))
	migDir := getEnv("MIGPATH")

	// to do: delegate to substitutions helper
	// mainSchema := getEnv("MAIN_SCHEMA")
	// appUser := getEnv("APPUSER")
	// appUserPwd := getEnv("APPUSER_PASSWORD")
	// guestUser := getEnv("GUESTUSER")
	// guestUserPwd := getEnv("GUESTUSER_PASSWORD")

	config := Config{
		PgUrl:         pgUrl,
		MigrationsDir: migDir,
	}

	return &config
}

func HydrateSQLTemplate(templateStr string, config Config) string {
	replacer := strings.NewReplacer(
	// TODO: delegate to substitutions
	// "${DB_NAME}", config.PgDb,
	// "${MAIN_SCHEMA}", config.MainSchema,
	// "${APPUSER}", config.AppUser,
	// "${APPUSER_PASSWORD}", config.AppUserPwd,
	)
	migrationStr := replacer.Replace(templateStr)

	return migrationStr
}
