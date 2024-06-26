package db

import "database/sql"

type PostgresDB struct {
	conn *sql.DB
}

func NewPostgredDB(connStr string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{conn: db}, err
}
