package core

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/netterminalmachine/nhml-graph/internal/helpers"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Migration struct {
	Id       int
	Filepath string
	Name     string
	Hash     string
}

func getFileContents(migrationFile string) (string, error) {
	byteArr, errFile := os.ReadFile(migrationFile)
	if errFile != nil {
		return "", fmt.Errorf("could not read migration file [%s]: %w", migrationFile, errFile)
	}
	return string(byteArr), nil
}

func allFiles(filesys fs.FS) (files []string, err error) {
	startFromRoot := "."
	err = fs.WalkDir(filesys, startFromRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("could not walk directory: %w", err)
		}

		if d.IsDir() {
			return nil // skip
		}

		files = append(files, path)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func makeHash(str string) (string, error) {
	byteArr := []byte(str)

	// could use bcrypt or similar for beefier sec but we are not fussed atm:
	hasher := sha1.New()
	_, hashErr := hasher.Write(byteArr)
	if hashErr != nil {
		return "", hashErr
	}

	// c/shouldve stored the integer hash, but for some reason I'd fancied storing the hex representation of the checksum and now the tbl is built that way, so:
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func getPendingMigrations(files []string, lastCommittedId int32) ([]Migration, error) {
	var latestId = int(lastCommittedId)
	var migs []Migration

	// sheer insouciance
	lenIdentifier := len("nnnn")

	for _, file := range files {
		len := len(file)
		if len < lenIdentifier {
			return nil, fmt.Errorf("cannot extract unique identifier from file [%s] with bad name", file)
		}
		id, errAtoi := strconv.Atoi(file[0:lenIdentifier])
		if errAtoi != nil {
			return nil, fmt.Errorf("cannot convert file prefix to integer: %w", errAtoi)
		}

		if id > latestId {
			_, filename := filepath.Split(file)
			migName := strings.Split(filename, ".")
			migs = append(migs, Migration{
				Id:       id,
				Filepath: file,
				Name:     migName[0][5:],
			})
		}
	}

	return migs, nil
}

func sqlForMigrationsRecord(mig Migration) (string, error) {
	if mig.Id == 0 || helpers.IsBlank(mig.Name) || helpers.IsBlank(mig.Hash) {
		slog.Error("invalid values",
			slog.Int("id", mig.Id),
			slog.String("name", mig.Name),
			slog.String("hash", mig.Hash),
		)
		return "", fmt.Errorf("need valid id, name and hash for migration")
	}

	sqlstr := fmt.Sprintf(
		"insert into migrations(id, name, hash) values (%d, '%s', '%s');",
		mig.Id, mig.Name, mig.Hash,
	)

	return sqlstr, nil
}

func getLatestCommittedMigrationId(ctx context.Context, pool *pgxpool.Pool) (int32, error) {
	var id int32
	var name string
	var hash string

	rows, err := pool.Query(ctx, "select id, name, hash from public.migrations order by id desc limit 1")
	if err != nil {
		return -1, fmt.Errorf("SQL query error: %w", err)
	}

	defer rows.Close()

	if rows.Next() {
		eScan := rows.Scan(&id, &name, &hash)
		if eScan != nil {
			return -1, fmt.Errorf("could not read returned rows: %w", err)
		}
		slog.Info("last migration",
			slog.Int("id", int(id)),
			slog.String("name", name),
			slog.String("hash", hash),
		)
		return id, nil
	}

	return 0, nil
}

func RunMigrations(ctx context.Context, config *helpers.Config, pool *pgxpool.Pool) error {
	fsys := os.DirFS(config.MigrationsDir)
	files, errFindAllFiles := allFiles(fsys)
	if errFindAllFiles != nil {
		return fmt.Errorf("cannot list migration files: %w", errFindAllFiles)
	}

	// prepare for migrations
	lastCommittedId, err := getLatestCommittedMigrationId(ctx, pool)
	if err != nil {
		return fmt.Errorf("could not get latest committed migration id: %w", err)
	}
	migs, err := getPendingMigrations(files, lastCommittedId)
	if err != nil {
		return fmt.Errorf("could not get pending migration files: %w", err)
	}

	if len(migs) == 0 {
		log.Println("No migrations to run.")
		return nil
	}

	err = asTransactionWithAutoRollback(ctx, pool, func(tx pgx.Tx) error {
		for _, mig := range migs {
			content, err := getFileContents(fmt.Sprintf("%s/%s", config.MigrationsDir, mig.Filepath))
			if err != nil {
				return err
			}
			var queries []string

			mig.Hash, err = makeHash(content)
			if err != nil {
				return err
			}
			queries = append(queries, content)

			sqlHashStore, err := sqlForMigrationsRecord(mig)
			if err != nil {
				return err
			}
			queries = append(queries, sqlHashStore)

			fingerprint := fmt.Sprintf("Migration for [%d][%s][%s]", mig.Id, mig.Name, mig.Hash)

			err = runSingleMigration(ctx, tx, queries)
			if err != nil {
				return err
			}
			slog.Info("✅ migration ok", slog.String("fingerprint", fingerprint))
		}

		return nil
	})

	return err
}

func CreateMigration(
	ctx context.Context,
	config *helpers.Config,
	pool *pgxpool.Pool,
	migName string,
) (string, error) {
	lastCommittedId, err := getLatestCommittedMigrationId(ctx, pool)
	if err != nil {
		slog.Error("error getting latest committed migration id", "error", err)
		return "", err
	}

	nextId := int(lastCommittedId) + 1
	cleanStr := helpers.SanitizeMigrationName(migName)

	targetName := fmt.Sprintf("%04d-%s.sql", nextId, cleanStr)
	targetPath := filepath.Join(config.MigrationsDir, targetName)

	path, err := filepath.Abs(targetPath)
	if err != nil {
		slog.Error("❌ Could not determine absolute file path", "error", err, "path", targetPath)
		return "", err
	}

	emptyBytArray := []byte("")
	err = os.WriteFile(path, emptyBytArray, 0644)
	if err != nil {
		slog.Error("❌ Could not write to file", "error", err, "path", targetPath)
		return "", err
	}
	slog.Info("✨ New migration file ready", slog.String("path", targetPath))
	return targetPath, nil
}
