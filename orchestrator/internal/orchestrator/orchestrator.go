package orchestrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"streaming/orchestrator/internal/db"
	"streaming/orchestrator/internal/streaming/pb"
)

var (
	ErrCannotCancelRequest = errors.New("cannot cancel request")
)

type Service struct {
	dbClient db.Client
	js       jetstream.JetStream
	logger   *zap.Logger

	pb.UnimplementedOrchestratorServer
}

func (r *Service) Run(ctx context.Context, in *pb.RunQuery) (*pb.RunResp, error) {
	job := &pb.Job{
		Id:     uuid.NewString(),
		State:  pb.State_STATE_INIT,
		Source: in.Source,
	}

	if err := r.dbClient.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	b, err := proto.Marshal(job)
	if err != nil {
		return nil, fmt.Errorf("marshal job: %w", err)
	}

	ack, err := r.js.Publish(ctx, "jobs", b)
	if err != nil {
		return nil, fmt.Errorf("publish job: %w", err)
	}

	r.logger.Info("publish job", zap.String("stream", ack.Stream), zap.Uint64("sequence", ack.Sequence))

	return &pb.RunResp{
		Id:    job.Id,
		State: job.State,
	}, nil
}

func (r *Service) Get(ctx context.Context, in *pb.GetQuery) (*pb.GetResp, error) {
	job, err := r.dbClient.ReadJob(ctx, in.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &pb.GetResp{
		Id:    job.Id,
		State: job.State,
	}, nil
}

func (r *Service) Cancel(ctx context.Context, in *pb.CancelQuery) (*pb.CancelResp, error) {
	job, err := r.dbClient.ReadJob(ctx, in.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	switch job.State {
	case pb.State_STATE_ACTIVE, pb.State_STATE_STARTUP, pb.State_STATE_INIT:
	default:
		return &pb.CancelResp{}, ErrCannotCancelRequest
	}

	b, err := proto.Marshal(nil) // TODO: Add protobuf to communicate with runner
	if err != nil {
		return nil, fmt.Errorf("marshal job: %w", err)
	}

	ack, err := r.js.Publish(ctx, "cancel", b)
	if err != nil {
		return nil, fmt.Errorf("publish cancel: %w", err)
	}

	r.logger.Info("publish cancel", zap.String("stream", ack.Stream), zap.Uint64("sequence", ack.Sequence))

	return &pb.CancelResp{
		Outcome: true,
	}, nil
}

func New(dbClient db.Client, js jetstream.JetStream, logger *zap.Logger) *Service {
	return &Service{
		dbClient: dbClient,
		js:       js,
		logger:   logger.Named("orchestrator"),
	}
}
