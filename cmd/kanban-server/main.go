package main

import (
	"bitbucket.org/latonaio/aion-core/cmd/kanban-server/app"
	"bitbucket.org/latonaio/aion-core/pkg/log"
)

func main() {
	log.SetFormat("status-kanban-server")
	env := app.GetConfig()
	// start server with microservices
	app.NewServer(env)
}
