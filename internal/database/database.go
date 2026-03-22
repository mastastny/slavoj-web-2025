package database

import (
	"context"
	"database/sql"
	_ "embed"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed init.sql
var initSQL string

func Init() *sql.DB {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "club.sqlite"
	}
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		panic(err)
	}

	if err := runMigrations(db); err != nil {
		panic(err)
	}

	return db
}

func runMigrations(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, initSQL)
	return err
}
