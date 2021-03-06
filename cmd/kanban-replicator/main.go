package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/latonaio/aion-core/cmd/kanban-replicator/app"
	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/services"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
	pb "bitbucket.org/latonaio/aion-core/proto/projectpb"
	"google.golang.org/grpc"
)

func main() {
	log.SetFormat("kanban-replicator")
	ctx := context.Background()
	env := app.GetConfig()

	redis := my_redis.NewRedisClient(env.GetRedisAddr())
	aionCh := make(chan *config.AionSetting)
	client := app.NewClient(ctx, env, redis)
	client.StartWatchKanban(ctx, aionCh)

	ya, err := config.LoadConfigFromDirectory(env.GetConfigPath())
	if err != nil {
		log.Fatal(err)

	}

	aionCh <- ya

	lis, err := net.Listen("tcp", ":11111")
	if err != nil {
		log.Fatalf("cant start server")
	}
	s := grpc.NewServer()
	server := services.NewProjectServer(aionCh, redis)
	pb.RegisterProjectServer(s, server)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("cant start server")
	}

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM)

loop:
	for {
		select {
		case s := <-signalCh:
			log.Printf("recieved signal: %s", s.String())
			ctx.Done()
			break loop
		}
	}
}
