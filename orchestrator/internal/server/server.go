package server

import (
	"streaming/orchestrator/internal/db"
	"streaming/orchestrator/internal/proto/pb"
)

type Server struct {
	hostPort     string
	dbClient     db.Client
	runnerClient pb.RunnerClient
}

func NewServer(hostPort string, dbClient db.Client) *Server {
	return &Server{
		hostPort: hostPort,
		dbClient: dbClient,
	}
}

func (r *Server) Start() {
}
