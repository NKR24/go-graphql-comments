package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	conn *sql.DB
}

func NewPostgresDB(connStr string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{conn: db}, nil
}
