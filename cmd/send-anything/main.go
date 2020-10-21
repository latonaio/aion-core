package main

import (
	"bitbucket.org/latonaio/aion-core/cmd/send-anything/app"
	"bitbucket.org/latonaio/aion-core/pkg/log"
)

func main() {
	log.SetFormat("send-anything")
	env := app.GetConfig()
	// start server with microservices
	app.NewServer(env)
}
