package client

import (
	"google.golang.org/grpc"

	"streaming/orchestrator/internal/proto/pb"
)

type OrchestratorClient interface {
	pb.RunnerClient
}

func NewOrchestrator(oURL string) (OrchestratorClient, error) {
	conn, err := grpc.NewClient(oURL)
	if err != nil {
		return nil, err
	}

	return pb.NewRunnerClient(conn), nil
}
