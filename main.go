package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	log "skhaz.dev/urlshortnen/logging"
	. "skhaz.dev/urlshortnen/pkg/functions"
	. "skhaz.dev/urlshortnen/pkg/router"
)

func main() {
	defer log.Sync()

	db, err := sql.Open("sqlite3", "/data/database.db")
	if err != nil {
		log.Fatal("failed to open the sqlite3 database", zap.Error(err))
	}
	defer db.Close()

	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user TEXT,
			count INTEGER DEFAULT 0,
			url TEXT UNIQUE,
			title TEXT,
			description TEXT,
			error TEXT DEFAULT NULL,
			ready INTEGER DEFAULT 0,
			updated_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user) REFERENCES users(email)
		)
	`); err != nil {
		log.Fatal("failed to create or ensure the 'data' table exists", zap.Error(err))
	}

	if _, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_data_url ON data (url)"); err != nil {
		log.Fatal("failed to create index on 'data(url)'", zap.Error(err))
	}

	if _, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_data_ready ON data (ready)"); err != nil {
		log.Fatal("failed to create index on 'data(ready)'", zap.Error(err))
	}

	if _, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_data_error ON data (error)"); err != nil {
		log.Fatal("failed to create index on 'data(error)'", zap.Error(err))
	}

	if _, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_data_created_at ON data (created_at)"); err != nil {
		log.Fatal("failed to create index on 'data(created_at)'", zap.Error(err))
	}

	if _, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_user ON data(user)"); err != nil {
		log.Fatal("failed to create index on 'data(user)'", zap.Error(err))
	}

	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			email TEXT PRIMARY KEY,
			password TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			active INTEGER DEFAULT 0
		)
	`); err != nil {
		log.Fatal("failed to create or ensure the 'users' table exists", zap.Error(err))
	}

	if _, err := db.Exec("UPDATE sqlite_sequence SET seq = 3000 WHERE name = 'data'"); err != nil {
		log.Fatal(err.Error())
	}

	go Synchronize(db)
	go Worker(db)
	go Collector(db)

	router := NewRouter(db)
	router.Start(":8000")
}
