package main

import (
	"os"

	"streaming/orchestrator/internal/controller/server"
	"streaming/orchestrator/internal/db"
)

func main() {
	hostPort := os.Getenv(server.EnvHostPort)
	sqlitePath := os.Getenv(db.EnvSQLite)
	sqliteClient := db.NewSQLiteClient(sqlitePath)
	srvr := server.NewServer(hostPort, sqliteClient)

	srvr.Start()
}
