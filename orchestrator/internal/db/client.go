package db

import (
	"context"

	"streaming/orchestrator/internal/streaming/pb"

	_ "github.com/lib/pq"
)

type Client interface {
	ReadJob(ctx context.Context, jobID string) (*pb.Job, error)
	CreateJob(ctx context.Context, job *pb.Job) error
}
