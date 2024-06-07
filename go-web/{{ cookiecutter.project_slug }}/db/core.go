package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/spf13/viper"
)

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
