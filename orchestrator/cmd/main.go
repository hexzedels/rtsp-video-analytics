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
	sqlitePath := os.Getenv(db.EnvSQLite)
	sqliteClient := db.MustNewSQLiteClient(sqlitePath, logger)

	nc, err := nats.Connect("")
	js, err := jetstream.New(nc)

	orchestratorService := orchestrator.New(sqliteClient, js, logger)

	server.New(host, port, sqliteClient, orchestratorService).Start()
}
