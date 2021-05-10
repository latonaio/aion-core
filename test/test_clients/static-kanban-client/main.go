package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/latonaio/aion-core/pkg/go-client/msclient"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
)

const msName = "static-kanban-client"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	client, err := msclient.NewKanbanClient(ctx, msName, kanbanpb.InitializeType_START_SERVICE)
	if err != nil {
		log.Fatalf("%v", err)
	}

	metadata := msclient.SetMetadata(
		map[string]interface{}{
			"result": "hoge",
		},
	)
	req, err := msclient.NewOutputData(metadata)
	err = client.OutputStaticKanban("test", req)

	recvCh := make(chan *kanbanpb.StaticKanban)
	go client.GetStaticKanban(ctx, "test", recvCh)

	for kanban := range recvCh {
		log.Printf("%#v", kanban.Id)
		log.Printf("%#v", kanban.StatusKanban)
		client.DeleteStaticKanban("test", kanban.Id)
	}

	quiteCh := make(chan os.Signal, 1)
	signal.Notify(quiteCh, syscall.SIGTERM)
	<-quiteCh
	cancel()
	return
}
