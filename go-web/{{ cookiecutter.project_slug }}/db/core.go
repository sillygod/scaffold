package db

import (
	"context"
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"ariga.io/atlas-go-sdk/atlasexec"
	"github.com/spf13/viper"
)

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

func NewPostgresqlDB() {

}

func NewSqliteDB(vp *viper.Viper) *sql.DB {
	name := vp.GetString("sqlite.db")

	db, err := sql.Open("sqlite3", name)
	if err != nil {
		panic(err)
	}

	return db
}
