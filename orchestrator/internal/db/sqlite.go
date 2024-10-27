package db

import (
	"context"
	"database/sql"
	"time"

	"streaming/orchestrator/internal/streaming/pb"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

const (
	createTableQuery = `CREATE TABLE IF NOT EXISTS jobs (
		id text PRIMARY KEY,
		state integer NOT NULL,
		source_url text NOT NULL,
		source_type integer NOT NULL
	) WITHOUT ROWID`

	createJobQuery = `INSERT INTO jobs (id, state, source_url, source_type) VALUES ($1, $2, $3, $4)`
	readJobQuery   = `SELECT * FROM jobs WHERE jobs.id = $1`
)

type SQLiteClient struct {
	db     *sql.DB
	logger *zap.Logger
}

func (r *SQLiteClient) ReadJob(ctx context.Context, jobID string) (*pb.Job, error) {
	var job pb.Job
	var source pb.Source

	err := r.db.QueryRowContext(ctx, readJobQuery, &jobID).Scan(
		&job.Id,
		&job.State,
		&source.Url,
		&source.SourceType,
	)
	if err != nil {
		return nil, err
	}

	job.Source = &source

	return &job, nil
}

func (r *SQLiteClient) CreateJob(ctx context.Context, job *pb.Job) error {
	_, err := r.db.ExecContext(
		ctx,
		createJobQuery,
		&job.Id,
		&job.State,
		&job.GetSource().Url,
		&job.GetSource().SourceType,
	)
	if err != nil {
		return err
	}

	return nil
}

func MustNewSQLiteClient(path string, logger *zap.Logger) *SQLiteClient {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if _, err := db.ExecContext(ctx, createTableQuery); err != nil {
		panic(err)
	}

	return &SQLiteClient{
		db:     db,
		logger: logger.Named("sqlite"),
	}
}
