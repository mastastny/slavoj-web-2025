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

	if _, err := db.Exec(`INSERT OR IGNORE INTO courts (id,name) VALUES (1,'Kurt #1')`); err != nil {
		panic(err)
	}

	if _, err := db.Exec(`INSERT OR IGNORE INTO courts (id,name) VALUES (2,'Kurt #2')`); err != nil {
		panic(err)
	}

	nowUTC := time.Now().UTC()
	startUTC := nowUTC.Truncate(time.Hour)
	endUTC := startUTC.Add(2 * time.Hour)

	if _, err := db.Exec(`
		INSERT OR IGNORE INTO reservations (court_id, start_at, end_at, name, email)
		VALUES (1, ?, ?, ?, ?)
		`,
		startUTC.Format(time.RFC3339),
		endUTC.Format(time.RFC3339),
		"Ukázková rezervace",
		"demo@example.com",
	); err != nil {
		panic(err)
	}

	if _, err := db.Exec(`
		INSERT OR IGNORE INTO reservations (court_id, start_at, end_at, name, email)
		VALUES (2, ?, ?, ?, ?)
		`,
		startUTC.Format(time.RFC3339),
		endUTC.Format(time.RFC3339),
		"Ukázková rezervace",
		"demo@example.com",
	); err != nil {
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
