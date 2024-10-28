package orchestrator

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"streaming/api/internal/streaming/pb"
)

type Client interface {
	pb.OrchestratorClient
}

func New(oURL string) (Client, error) {
	conn, err := grpc.NewClient(oURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return pb.NewOrchestratorClient(conn), nil
}
