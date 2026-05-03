package migrations

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/AnxVit/metrics-server/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
	dir   = flags.String("dir", ".", "directory with migration files")
)

//go:embed pgmigrations/*.sql
var migrationFiles embed.FS

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

	arguments := []string{}
	if len(args) > 2 {
		arguments = append(arguments, args[2:]...)
	}

	Migrate(dbstring, command, arguments)
}

func Migrate(dbstring, command string, arguments []string) {
	db, err := goose.OpenDBWithDriver("postgres", dbstring)
	if err != nil {
		logger.Log.Warn("goose: failed to open DB", zap.Error(err))
		return
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Log.Warn("goose: failed to close DB:", zap.Error(err))
		}
	}()

	goose.SetBaseFS(migrationFiles)

	ctx := context.Background()
	if err := goose.RunContext(ctx, command, db, "pgmigrations", arguments...); err != nil {
		logger.Log.Warn(fmt.Sprintf("goose %v", command), zap.Error(err))
	}
}
