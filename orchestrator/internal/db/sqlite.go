package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteClient struct {
	*baseClient
}

func NewSQLiteClient(path string) *SQLiteClient {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		panic(err)
	}

	return &SQLiteClient{
		newBaseClient(db),
	}
}
