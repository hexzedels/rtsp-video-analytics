package orchestrator

import "streaming/orchestrator/internal/streaming/pb"

type Service struct {
	pb.UnimplementedOrchestratorServer
}

func New() *Service {
	return &Service{}
}
