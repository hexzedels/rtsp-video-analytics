package main

import (
	"os"

	"streaming/api/internal/config"
	"streaming/api/internal/controller/server"
)

func main() {
	hostPort := os.Getenv(config.EnvHostPort)
	orchestratorURL := os.Getenv(config.EnvOrchestratorURL)

	srvr := server.NewServer(hostPort).SetOrchestratorClient(orchestratorURL)

	srvr.Start()
}
