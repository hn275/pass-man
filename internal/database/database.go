package database

import (
	"log"

	"github.com/hn275/pass-man/internal/paths"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	conn *sqlx.DB
)

func init() {
	var err error
	dbPath := paths.MakePath("database.db")
	conn, err = sqlx.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	err = conn.Ping()
	if err != nil {
		log.Fatal(err)
	}

	migration(conn)
}

func New() *sqlx.DB {
	return conn
}

func migration(db *sqlx.DB) {
	schemas := `
    CREATE TABLE IF NOT EXISTS pass (
    id TEXT NOT NULL PRIMARY KEY,
    user TEXT NOT NULL,
    pass TEXT NOT NULL,
    site TEXT NOT NULL
    );
    `
	db.MustExec(schemas)
}
