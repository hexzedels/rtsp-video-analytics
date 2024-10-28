package server

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"streaming/api/internal/controller/handler/job"
	"streaming/api/internal/pkg/orchestrator"
)

type Server struct {
	hostPort   string
	orchClient orchestrator.Client
	logger     *zap.Logger
}

func NewServer(logger *zap.Logger, hostPort string) *Server {
	return &Server{
		hostPort: hostPort,
		logger:   logger.Named("api"),
	}
}

func (r *Server) SetOrchestratorClient(oURL string) *Server {
	orchClient, err := orchestrator.New(oURL)
	if err != nil {
		panic(err)
	}

	r.orchClient = orchClient

	return r
}

func (r *Server) Start() {
	api := r.newAPI()

	api.Run(r.hostPort)
}

func (r *Server) newAPI() *gin.Engine {
	eng := gin.New()

	apiV1 := eng.Group("/v1")

	privateRouter := NewPrivate(r.orchClient, r.logger)
	job.AttachToGroup(apiV1.Group("/jobs"), privateRouter)

	return eng
}
