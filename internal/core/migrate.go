package core

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io/fs"
	"log"
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
	Type     string
	Hash     string
}

func getFileContents(migrationFile string) string {
	byteArr, errFile := os.ReadFile(migrationFile)
	if errFile != nil {
		log.Fatal("Failed to read file!")
	}
	return string(byteArr)
}

func allFiles(filesys fs.FS) (files []string, err error) {
	startFromRoot := "."
	if err := fs.WalkDir(filesys, startFromRoot, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func makeHash(str string) string {
	byteArr := []byte(str)

	// could use bcrypt or similar for beefier sec but we are not fussed atm:
	hasher := sha1.New()
	_, hashErr := hasher.Write(byteArr)
	if hashErr != nil {
		log.Fatal("can't hash the migration!")
	}

	// c/shouldve stored the integer hash, but for some reason I'd fancied storing the hex representation of the checksum and now the tbl is built that way, so:
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func getPendingMigrations(files []string, lastCommittedId int32) []Migration {
	var latestId = int(lastCommittedId)
	var migs []Migration

	// sheer insouciance
	lenIdentifier := len("nnnn")

	for _, file := range files {
		len := len(file)
		if len < lenIdentifier {
			log.Fatalf("Encountered file [%s] with bad name - cannot extract unique identifier for migration. Bailing.", file)
		}
		id, _ := strconv.Atoi(file[0:lenIdentifier])

		if id > latestId {
			_, filename := filepath.Split(file)
			migName := strings.Split(filename, ".")
			migType := "core"
			if strings.EqualFold(filepath.Ext(filename), ".json") {
				migType = "authz"
			}

			migs = append(migs, Migration{
				Id:       id,
				Filepath: file,
				Name:     migName[0][5:],
				Type:     migType,
			})
		}
	}

	return migs
}

func sqlForMigrationsRecord(mig Migration) string {
	if mig.Id == 0 || helpers.IsBlank(mig.Name) || helpers.IsBlank(mig.Hash) {
		log.Fatalf("Need valid id, name and hash for migration!")
	}

	return fmt.Sprintf("insert into migrations(id, name, mig_type, hash) values (%d, '%s', '%s', '%s');", mig.Id, mig.Name, mig.Type, mig.Hash)
}

func runInTransaction(ctx context.Context, tx pgx.Tx, sqlStrings []string, ch chan bool) {
	allOK := true

	for _, sql := range sqlStrings {
		_, err := tx.Exec(ctx, sql)
		if err != nil {
			fmt.Printf("SQL transaction error: %v\n", err)
			allOK = false
			break
		}
	}

	ch <- allOK
}

func getLatestCommittedMigrationId(ctx context.Context, pool *pgxpool.Pool, ch chan int32) {
	var id int32
	var name string
	var hash string

	rows, err := pool.Query(ctx, "select id, name, hash from public.migrations order by id desc limit 1")
	if err != nil {
		log.Fatalf("SQL query error: %v\n", err)
	}

	defer rows.Close()

	if rows.Next() {
		eScan := rows.Scan(&id, &name, &hash)
		if eScan != nil {
			log.Fatalf("Could not read returned rows! %v\n", err)
		}
		fmt.Printf("🔒 Last migration: [%d][%s][%s]\n", id, name, hash)
		ch <- id
		return
	}

	ch <- 0
}

func RunMigration(ctx context.Context, config *helpers.Config, pool *pgxpool.Pool) {
	fsys := os.DirFS(config.MigrationsDir)
	files, errFindAllFiles := allFiles(fsys)
	if errFindAllFiles != nil {
		log.Fatal("can't list migration files!")
	}

	ch := make(chan int32)
	chmig := make(chan bool)

	// prepare for migrations
	go getLatestCommittedMigrationId(ctx, pool, ch)
	lastCommittedId := <-ch
	migs := getPendingMigrations(files, lastCommittedId)

	var ok bool
	var fingerprint string
	for _, mig := range migs {
		tx, _ := pool.Begin(ctx)
		content := getFileContents(fmt.Sprintf("%s/%s", config.MigrationsDir, mig.Filepath))
		var queries []string

		if mig.Type == "core" {
			sqlStr := helpers.HydrateSQLTemplate(content, *config)
			mig.Hash = makeHash(sqlStr)
			queries = append(queries, sqlStr)
		} else {
			// mig.Type authz
			mig.Hash = makeHash(content)
		}

		sqlHashStore := sqlForMigrationsRecord(mig)
		queries = append(queries, sqlHashStore)

		fingerprint = fmt.Sprintf("Migration for [%d][%s][%s]", mig.Id, mig.Name, mig.Hash)

		if mig.Type == "core" {
			go runInTransaction(ctx, tx, queries, chmig)
			ok = <-chmig
		} else {
			fmt.Println("Migrating metadata is not supported yet.")
		}

		if ok {
			fmt.Printf("✅ Committing: %s\n", fingerprint)
			eTrxn := tx.Commit(ctx)
			if eTrxn != nil {
				log.Panicln("❌Dagnammit... the commit transaction failed!! 😫")
				ok = false
			}
		} else {
			fmt.Printf("❌ Failed - Rolling Back: %s\n", fingerprint)
			eTrxn := tx.Rollback(ctx)
			if eTrxn != nil {
				log.Panicln("💣Uhm... so... the rollback ALSO failed. OMG. 💀")
			}
			break //stop processing
		}
	}

	if !ok {
		if len(migs) == 0 {
			log.Println("No migrations to run.")
		}
		log.Fatalf("Stopping...")
	}
}

func CreateMigration(
	ctx context.Context,
	config *helpers.Config,
	pool *pgxpool.Pool,
	migName string,
	migType string,
) string {
	ch := make(chan int32)

	go getLatestCommittedMigrationId(ctx, pool, ch)
	lastCommittedId := <-ch

	nextId := int(lastCommittedId) + 1

	cleanStr := helpers.SanitizeMigrationName(migName)

	// identify regular sql migrations vs metadata migrations controlled by other subsystems
	ext := "sql"
	if migType != "core" {
		ext = "json"
	}
	targetName := fmt.Sprintf("%04d-%s.%s", nextId, cleanStr, ext)
	targetPath := filepath.Join(config.MigrationsDir, targetName)

	p, e := filepath.Abs(targetPath)
	if e != nil {
		log.Fatalf("❌ Could not determine absolute file path! %v", e)
	}

	emptyBytArray := []byte("")
	eWrite := os.WriteFile(p, emptyBytArray, 0644)
	if eWrite != nil {
		log.Fatalf("❌ Could not write to file! %v", eWrite)
	}
	fmt.Printf("✨ New migration file ready: %s\n", targetPath)
	return targetPath
}
