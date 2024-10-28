package server

import (
	"go.uber.org/zap"

	"streaming/api/internal/pkg/orchestrator"
)

type Private struct {
	orchestratorClient orchestrator.Client
	logger             *zap.Logger
}

func (r *Private) OrchestratorClient() orchestrator.Client {
	return r.orchestratorClient
}

func (r *Private) Logger() *zap.Logger {
	return r.logger
}

func NewPrivate(oClient orchestrator.Client, logger *zap.Logger) *Private {
	return &Private{
		orchestratorClient: oClient,
		logger:             logger,
	}
}
