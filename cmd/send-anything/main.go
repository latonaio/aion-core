package main

import (
	"bitbucket.org/latonaio/aion-core/cmd/send-anything/app"
	"bitbucket.org/latonaio/aion-core/pkg/log"
)

const buildDate = "14:17"

func main() {
	log.SetFormat("send-anything")
	log.Printf("build by %v", buildDate)
	env := app.GetConfig()
	// start server with microservices
	app.NewServer(env)
}
