package main

import (
	"os"
	"strconv"

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

	sqlitePath := os.Getenv(db.EnvSQLite)
	sqliteClient := db.NewSQLiteClient(sqlitePath)
	orchestratorService := orchestrator.New()
	server.New(host, port, sqliteClient, orchestratorService).Start()
}
