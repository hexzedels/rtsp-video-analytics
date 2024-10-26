package server

import (
	"go.uber.org/zap"

	"streaming/orchestrator/pkg/client"
)

type Private struct {
	orchestratorClient client.OrchestratorClient
	logger             *zap.Logger
}

func (r *Private) OrchestratorClient() client.OrchestratorClient {
	return r.orchestratorClient
}

func (r *Private) Logger() *zap.Logger {
	return r.logger
}

func NewPrivate(oClient client.OrchestratorClient) *Private {
	return &Private{orchestratorClient: oClient}
}
