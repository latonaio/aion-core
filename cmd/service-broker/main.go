// Copyright (c) 2019-2020 Latona. All rights reserved.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/latonaio/aion-core/cmd/service-broker/app"
	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/k8s"
	"bitbucket.org/latonaio/aion-core/pkg/log"
)

func main() {
	ctx := context.Background()
	log.SetFormat("service-broker")
	env := app.GetConfig()

	if err := config.GetInstance().LoadConfig(env.GetConfigPath(), env.IsDocker()); err != nil {
		log.Fatalf("cant open yaml file (path: %s): %v", env.GetConfigPath(), err)
	}

	if env.IsDocker() {
		log.Printf("Use docker mode")
		if err := k8s.New(
			ctx, env.GetDataDir(), env.GetRepositoryPrefix(), env.GetNamespace(), env.GetRegistrySecret()); err != nil {
			log.Fatal(err)
		}
	}

	msc, err := app.StartMicroservicesController(ctx, env)
	if err != nil {
		log.Fatalf("cant start microservice controller: %v", err)
	}

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM)

loop:
	for {
		select {
		case s := <-signalCh:
			log.Printf("recieved signal: %s", s.String())
			msc.StopAllMicroservice()
			break loop
		}
	}
}
