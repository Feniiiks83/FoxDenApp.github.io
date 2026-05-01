package storage

import (
	"database/sql"
	"fmt"
	"foxDenApp/internal/config"

	_ "github.com/lib/pq"
)

type Storage struct {
	Db *sql.DB
}

func Connect(cfg *config.Config) *sql.DB {
	var connString = fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPass, cfg.DbName,
	)

	db, err := sql.Open("postgres", connString)

	if err != nil {
		panic(fmt.Sprintf("Couldn't connect to database! Error: %s", err.Error()))
	}

	return db
}

func (storage *Storage) Prepare() {
	_, err := storage.Db.Exec(
		`
		CREATE TABLE IF NOT EXISTS users
		(
			id SERIAL PRIMARY KEY,
			time TIMESTAMPTZ Not NULL,
			ip character varying(15)
		)	
		`,
	)

	if err != nil {
		panic(fmt.Sprintf("Couldn't create users table! Error: %s", err.Error()))
	}
}

func Init(db *sql.DB) *Storage {
	storage := Storage{
		Db: db,
	}

	return &storage
}
