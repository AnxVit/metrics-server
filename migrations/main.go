package main

import (
	"context"
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
	dir   = flags.String("dir", ".", "directory with migration files")
)

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("goose: failed to parse flags: %v", err)
	}
	args := flags.Args()

	if len(args) < 3 {
		flags.Usage()
		return
	}

	command, dbstring := args[0], args[1]

	migrationDir := *dir
	if len(args) > 2 {
		migrationDir = args[2]
	}

	db, err := goose.OpenDBWithDriver("postgres", dbstring)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v", err)
		}
	}()

	arguments := []string{}
	if len(args) > 3 {
		arguments = append(arguments, args[3:]...)
	}

	ctx := context.Background()
	if err := goose.RunContext(ctx, command, db, migrationDir, arguments...); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
}
