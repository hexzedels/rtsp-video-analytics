package main

import (
	"os"
	"strconv"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"streaming/orchestrator/internal/controller/server"
	"streaming/orchestrator/internal/db"
	"streaming/orchestrator/internal/orchestrator"
)

func main() {
	streamName := os.Getenv(server.EnvNatsStream)
	host := os.Getenv(server.EnvHost)
	portRaw := os.Getenv(server.EnvPort)
	if portRaw == "" {
		panic("empty port")
	}

	port, err := strconv.Atoi(portRaw)
	if err != nil {
		panic(err)
	}

	logger, err := zap.NewProduction(zap.AddStacktrace(zapcore.ErrorLevel), zap.AddCaller())
	if err != nil {
		panic(err)
	}

	sqlitePath := os.Getenv(server.EnvSQLite)
	sqliteClient := db.MustNewSQLiteClient(sqlitePath, logger)

	natsURL := os.Getenv(server.EnvNatsURL)
	nc, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		panic(err)
	}

	orchestratorService := orchestrator.New(sqliteClient, js, logger, streamName)

	server.New(host, port, sqliteClient, orchestratorService).Start()
}
