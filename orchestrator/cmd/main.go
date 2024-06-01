package main

import (
	"os"
	"streaming/orchestrator/internal/db"
	"streaming/orchestrator/internal/server"
)

func main() {
	hostPort := os.Getenv(server.EnvHostPort)
	sqlitePath := os.Getenv(db.EnvSQLite)
	sqliteClient := db.NewSQLiteClient(sqlitePath)
	srvr := server.NewServer(hostPort, sqliteClient)

	srvr.Start()
}
