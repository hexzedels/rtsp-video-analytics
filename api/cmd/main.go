package main

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"streaming/api/internal/config"
	"streaming/api/internal/controller/server"
)

func main() {
	hostPort := os.Getenv(config.EnvHostPort)
	orchestratorURL := os.Getenv(config.EnvOrchestratorURL)

	logger, err := zap.NewProduction(zap.AddStacktrace(zapcore.ErrorLevel), zap.AddCaller())
	if err != nil {
		panic(err)
	}

	srvr := server.NewServer(logger, hostPort).SetOrchestratorClient(orchestratorURL)

	srvr.Start()
}
