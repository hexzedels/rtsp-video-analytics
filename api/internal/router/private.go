package router

import "streaming/api/internal/db"

type Private struct {
	dbClient db.Client
}

func NewPrivate(dbClient db.Client) *Private {
	return &Private{dbClient: dbClient}
}
