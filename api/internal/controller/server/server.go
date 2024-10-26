package server

import (
	"github.com/gin-gonic/gin"

	"streaming/api/internal/controller/handler/job"
	"streaming/orchestrator/pkg/client"
)

type Server struct {
	hostPort   string
	orchClient client.OrchestratorClient
}

func NewServer(hostPort string) *Server {
	return &Server{
		hostPort: hostPort,
	}
}

func (r *Server) SetOrchestratorClient(oURL string) *Server {
	orchClient, err := client.NewOrchestrator(oURL)
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

	privateRouter := NewPrivate(r.orchClient)
	job.AttachToGroup(apiV1.Group("/jobs"), privateRouter)

	return eng
}
