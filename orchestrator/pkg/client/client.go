package client

import (
	"google.golang.org/grpc"

	"streaming/orchestrator/internal/streaming/pb"
)

type OrchestratorClient interface {
	pb.OrchestratorClient
}

func NewOrchestrator(oURL string) (OrchestratorClient, error) {
	conn, err := grpc.NewClient(oURL)
	if err != nil {
		return nil, err
	}

	return pb.NewOrchestratorClient(conn), nil
}
