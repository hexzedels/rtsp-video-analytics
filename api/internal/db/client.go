package db

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Client interface {
	Read(ctx context.Context, query *Query) (*sql.Rows, error)
	Create(ctx context.Context, payload *Payload) error
}

type Query struct {
	Payload string
	Args    []any
}

type Payload struct {
	Resource any
	Query    string
	Args     []any
}

type baseClient struct {
	db *sql.DB
}

func (r *baseClient) Read(ctx context.Context, f *Query) (*sql.Rows, error) {
	rows, err := r.db.QueryContext(ctx, f.Payload, f.Args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *baseClient) Create(ctx context.Context, payload *Payload) error {
	_, err := r.db.ExecContext(ctx, payload.Query, payload.Args...)
	if err != nil {
		return err
	}

	return nil
}

func newBaseClient(db *sql.DB) *baseClient {
	return &baseClient{db: db}
}
