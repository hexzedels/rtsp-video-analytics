package main

import (
	"os"
	"streaming/api/internal/db"
	"streaming/api/internal/server"
)

func main() {
	hostPort := os.Getenv(server.EnvHostPort)
	sqlitePath := os.Getenv(db.EnvSQLite)
	sqliteClient := db.NewSQLiteClient(sqlitePath)
	srvr := server.NewServer(hostPort, sqliteClient)

	srvr.Start()
}
