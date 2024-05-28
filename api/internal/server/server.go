package server

import (
	"streaming/api/internal/db"
	"streaming/api/internal/handlers"
	"streaming/api/internal/router"

	"github.com/gin-gonic/gin"
)

type Server struct {
	hostPort string
	dbClient db.Client
}

func NewServer(hostPort string, dbClient db.Client) *Server {
	return &Server{
		hostPort: hostPort,
		dbClient: dbClient,
	}
}

func (r *Server) Start() {
	api := r.newAPI()

	api.Run(r.hostPort)
}

func (r *Server) newAPI() *gin.Engine {
	eng := gin.New()

	apiV1 := eng.Group("/api/v1")

	privateRouter := router.NewPrivate(r.dbClient)
	handlers.AttachToGroup(apiV1.Group("/job"), privateRouter)

	return eng
}
