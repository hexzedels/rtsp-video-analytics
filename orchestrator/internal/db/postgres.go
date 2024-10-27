package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type PostgresClient struct {
	db *sql.DB
}

func NewPostgresClient(uri string) *PostgresClient {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		panic(err)
	}

	return &PostgresClient{
		db: db,
	}
}
