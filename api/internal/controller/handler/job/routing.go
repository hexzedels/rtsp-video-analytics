package job

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"streaming/orchestrator/pkg/client"
)

type jobPrivateRouter struct {
	orchestratorClient client.OrchestratorClient
	logger             *zap.Logger
}

type clientProvider interface {
	OrchestratorClient() client.OrchestratorClient
	Logger() *zap.Logger
}

func AttachToGroup(group *gin.RouterGroup, cli clientProvider) {
	jobRouter := jobPrivateRouter{
		orchestratorClient: cli.OrchestratorClient(),
		logger:             cli.Logger(),
	}

	group.POST("", jobRouter.new)
	group.GET("/:id", jobRouter.get)
	group.POST("/:id", jobRouter.post)
}
