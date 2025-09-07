package main

import (
	"context"
	"database/sql"
	"io"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Event struct {
	Title string    `json:"title"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	// případně: ID, Color, extendedProps…
}

func Init() *sql.DB {
	// 1) Otevři DB (soubor vedle binárky; přepni dle potřeby)
	db, err := sql.Open("sqlite3", "file:club.sqlite?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		panic(err)
	}

	// 2) Spusť migraci
	if err := runMigrations(db, "init.sql"); err != nil {
		panic(err)
	}

	// 3) Seed – prvni kurt
	if _, err := db.Exec(`INSERT OR IGNORE INTO courts (id,name) VALUES (1,'Kurt #1')`); err != nil {
		panic(err)
	}

	// 3) Seed – druhy kurt
	if _, err := db.Exec(`INSERT OR IGNORE INTO courts (id,name) VALUES (2,'Kurt #2')`); err != nil {
		panic(err)
	}
	// seed ukázkové rezervace: [aktuální_hodina_start, další_hodina_end] => 2h blok
	nowUTC := time.Now().UTC()
	startUTC := nowUTC.Truncate(time.Hour) // začátek aktuální hodiny (UTC)
	endUTC := startUTC.Add(2 * time.Hour)  // konec následující hodiny (UTC)

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

func runMigrations(db *sql.DB, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sqlBytes, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	// Spustíme v jedné transakci
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, string(sqlBytes))
	return err
}
