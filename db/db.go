package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func Init(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task TEXT NOT NULL,
		completed INTEGER NOT NULL DEFAULT 0,
		priority INTEGER NOT NULL DEFAULT 2,
		project TEXT,
		due_date TEXT,
		recurring TEXT,
		next_due TEXT,
		created_at TEXT NOT NULL,
		completed_at TEXT
	);
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}
