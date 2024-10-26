package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"streaming/orchestrator/internal/db"
	"streaming/orchestrator/internal/orchestrator"
	"streaming/orchestrator/internal/streaming/pb"
)

type Server struct {
	host       string
	port       int
	dbClient   db.Client
	grpcServer *grpc.Server
	logger     *zap.Logger
}

func New(host string, port int, dbClient db.Client, orchestrator *orchestrator.Service) *Server {
	s := &Server{
		host:     host,
		port:     port,
		dbClient: dbClient,
	}

	s.grpcServer = grpc.NewServer()

	pb.RegisterOrchestratorServer(s.grpcServer, orchestrator)

	return s
}

func (r *Server) Start() {
	go r.startGrpc(r.port)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	r.logger.Info("gracefully stopped server")
	os.Exit(0)
}

func (r *Server) startGrpc(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	if err := r.grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
