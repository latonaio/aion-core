// Copyright (c) 2019-2020 Latona. All rights reserved.
package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/latonaio/aion-core/cmd/service-broker/app"
	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/services"
	"bitbucket.org/latonaio/aion-core/pkg/k8s"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	pb "bitbucket.org/latonaio/aion-core/proto/projectpb"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	log.SetFormat("service-broker")
	env := app.GetConfig()

	ya, err := config.LoadConfigFromDirectory(env.GetConfigPath(), env.IsDocker())
	if err != nil {
		log.Fatalf("cant open yaml file (path: %s): %v", env.GetConfigPath(), err)
	}

	if env.IsDocker() {
		log.Printf("Use docker mode")
		if err := k8s.New(
			ctx, env.GetDataDir(), env.GetRepositoryPrefix(), env.GetNamespace(), env.GetRegistrySecret()); err != nil {
			log.Fatal(err)
		}
	}

	aionCh := make(chan *config.AionSetting)
	msc, err := app.StartMicroservicesController(ctx, env, aionCh)
	if err != nil {
		log.Fatalf("cant start microservice controller: %v", err)
	}

	aionCh <- ya

	lis, err := net.Listen("tcp", ":11111")
	if err != nil {
		log.Fatalf("cant start server")
	}
	s := grpc.NewServer()
	server := &services.ProjectServer{AionCh: aionCh, IsDocker: env.IsDocker()}
	pb.RegisterProjectServer(s, server)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("cant start server")
	}

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM)

	for {
		select {
		case s := <-signalCh:
			log.Printf("recieved signal: %s", s.String())
			msc.StopAllMicroservice()
			goto END
		}
	}
END:
}
