package db

import (
	"exampleproj/config"
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/fx"

	"ariga.io/atlas-go-sdk/atlasexec"
)

// Migrate applies the migrations for the application
func Migrate(dsn string) (*atlasexec.MigrateApply, error) {
	workdir, err := atlasexec.NewWorkingDir(atlasexec.WithMigrations(os.DirFS("./db/migrations/")))
	if err != nil {
		return nil, err
	}

	defer workdir.Close()

	// Initialize the client.
	client, err := atlasexec.NewClient(workdir.Path(), "atlas")
	if err != nil {
		return nil, err
	}
	// Run `atlas migrate apply` on a SQLite database under /tmp.
	res, err := client.MigrateApply(context.Background(), &atlasexec.MigrateApplyParams{
		URL: dsn,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func GetPostgresqlDSN(config *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		config.DB.USER,
		config.DB.PASSWORD,
		config.DB.HOST,
		config.DB.PORT,
		config.DB.NAME,
	)
}

func NewPostgresqlDB(lc fx.Lifecycle, config *config.Config) *pgx.Conn {
	ctx := context.Background()
	dsn := GetPostgresqlDSN(config)
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		panic(err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error { return conn.Close(ctx) },
	})

	return conn
}

func NewSqliteDB(config *config.Config) *sql.DB {
	name := config.DB.NAME
	db, err := sql.Open("sqlite3", name)
	if err != nil {
		panic(err)
	}

	return db
}
