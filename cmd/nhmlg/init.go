package main

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/urfave/cli/v2"

	"github.com/netterminalmachine/nhml-graph/internal/helpers"
)

func InitialiseMigrations(ctx *cli.Context) error {
	if ctx == nil {
		return errors.New("parent context is missing")
	}

	config := helpers.LoadConfig()
	sqlStr := "create table if not exists migrations (id int primary key, name text not null, mig_type text not null, hash text not null);"

	conn, err := pgx.Connect(ctx.Context, config.PgUrl)

	if err != nil {
		return err
	}

	defer func() {
		ctxCloseErr := conn.Close(ctx.Context)
		if ctxCloseErr != nil {
			fmt.Printf("context close error: %v\n", ctxCloseErr)
		}
	}()

	statusText, errTx := conn.Exec(ctx.Context, sqlStr)
	if errTx != nil {
		fmt.Printf("SQL transaction error: %v\nStatus text:%s\n", errTx, statusText)
		return errTx
	}

	fmt.Println("Migrations table ready.")

	return nil
}
