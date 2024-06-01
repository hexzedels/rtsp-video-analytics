package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type PostgresClient struct {
	*baseClient
}

func NewPostgresClient(uri string) *PostgresClient {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		panic(err)
	}

	return &PostgresClient{
		newBaseClient(db),
	}
}
