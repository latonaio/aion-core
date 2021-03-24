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
	clspb "bitbucket.org/latonaio/aion-core/proto/clusterpb"
	pjpb "bitbucket.org/latonaio/aion-core/proto/projectpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	var (
		err error
		ya  *config.AionSetting
	)
	aionCh := make(chan *config.AionSetting, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.SetFormat("service-broker")
	env := app.GetConfig()

	if env.IsDefaultMode() {
		ya, err = config.LoadConfigFromDirectory(env.GetConfigPath(), env.IsDocker())
		if err != nil {
			log.Fatalf("cant open yaml file (path: %s): %v", env.GetConfigPath(), err)
		}
		log.Debugf("[default mode]load config from yaml: %+v \n", ya)
		aionCh <- ya
	}
	log.Debugln("debug aion chan <- done ")
	if env.IsDocker() {
		log.Printf("Use docker mode")
		if err := k8s.New(
			ctx, env.GetDataDir(), env.GetRepositoryPrefix(), env.GetNamespace(), env.GetRegistrySecret()); err != nil {
			log.Fatal(err)
		}
	}

	// service deploy　controller
	msc, err := app.StartMicroservicesController(ctx, env, aionCh)
	if err != nil {
		log.Fatalf("cant start microservice controller: %v", err)
	}
	log.Println("start MicroservicesController")

	switch env.GetMode() {
	case app.MasterMode:
		workerStatusMonitoringCh := make(chan map[string]map[string]bool)
		log.Println("start ServiceBroker MasterServer ")
		// start grpc server
		go masterServer(workerStatusMonitoringCh, env)
		// start worker status monitor
		go app.NewWorkerStatusMonitor(workerStatusMonitoringCh).Start()
	case app.WorkerMode:
		// start grpc client
		go workerClient(ctx, env, msc, aionCh)
	}

	log.Println("started all process")
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM)

	for {
		s := <-signalCh
		log.Printf("received signal: %s", s.String())
		msc.StopAllMicroservice()
		return
	}
}

func masterServer(workerStatusMonitoringCh chan<- map[string]map[string]bool, env *app.Config) {
	sendAionSettingToWorkerCh := make(chan *config.AionSetting, 1)
	lis, err := net.Listen("tcp", ":11111")
	if err != nil {
		log.Fatalf("cant start server")
	}

	s := grpc.NewServer()
	ww := services.NewWorkerWatcher(workerStatusMonitoringCh)
	server := &services.ProjectServer{AionCh: sendAionSettingToWorkerCh, IsDocker: env.IsDocker()}
	clspb.RegisterClusterServer(s, ww)
	pjpb.RegisterProjectServer(s, server)

	//　deploy指示をworkerへ送信
	go ww.SendAionSettingToWorker(sendAionSettingToWorkerCh)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("cant start server")
	}
}

func workerClient(ctx context.Context, env *app.Config, msc mscStatus, applyAionSettingCh chan<- *config.AionSetting) {
	conn, err := grpc.DialContext(ctx, "aion-servicebroker.master.svc.cluster.local:11110", grpc.WithInsecure())
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("[worker] grpc conn close failed cause :%v", err)
		}
	}()
	if err != nil {
		log.Fatalf("grpc dial failed , cause: %v", err)
	}

	client := clspb.NewClusterClient(conn)
	meta := &clspb.NodeMeta{NodeIP: env.GetNodeIP(), NodeName: env.GetNodeName()}
	grpcMd := metadata.Pairs("NodeName", meta.NodeName, "NodeIP", meta.NodeIP)
	mdCtx := metadata.NewOutgoingContext(context.Background(), grpcMd)
	stream, err := client.JoinMasterAion(mdCtx)
	if err != nil {
		log.Fatalf("grpc JoinMasterAion request failed ,cause:", err)
	}
	log.Println("start ServiceBroker workerClient ")
	if err := stream.Send(meta); err != nil {
		log.Fatalf("[gcp][client]send failed cause: %v", err)
	}

	log.Debugf("joined to master, who am i => %+v", meta)
	updateTrigger := msc.GetStatusUpdateTrigger()

	// master　からのデプロイ指示
	go func() {
		for {
			rs := new(clspb.Apply)
			if err := stream.RecvMsg(rs); err != nil {
				log.Fatalf("grpc stream RecvMsg failed ,cause:", err)
			}
			log.Debugf("[grpc][client] RecvMsg service setting from master: %+v", rs)

			aionSetting, err := config.LoadConfigFromGRPC(rs.AionSetting, env.IsDocker())
			if err != nil {
				log.Printf("[worker][workerClient][LoadConfigFromGRPC] failed cause:%v", err)
			}

			applyAionSettingCh <- aionSetting
		}
	}()

	// worker上のservices状態の更新
	for {
		// 状態更新された
		updatedStatus := <-updateTrigger
		meta.ServicesStatus = updatedStatus
		if err := stream.Send(meta); err != nil {
			log.Fatalf("grpc stream send failed ,cause:", err)
		}

		log.Println("debug grpc client requested")
	}
}

type mscStatus interface {
	GetStatusUpdateTrigger() <-chan map[string]bool
	GetMicroServicesStatus() map[string]bool
}
