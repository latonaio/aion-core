package main

import (
	"context"
	"log"
	"sync"
	"syscall"

	"bitbucket.org/latonaio/aion-core/pkg/go-client/msclient"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
)

const msName = "aion-send-test-kube"

var whitelist = []string{
	"ContainerCreating",
	"PodInitializing",
}

var countList = sync.Pool{
	New: func() interface{} {
		return map[string]int{}
	},
}

func main() {
	errCh := make(chan error, 1)
	quiteCh := make(chan syscall.Signal, 1)

	ctx, cancel := context.WithCancel(context.Background())

	c, err := msclient.NewKanbanClient(ctx)
	if err != nil {
		errCh <- err
	}

	_, err = c.SetKanban(msName, c.GetProcessNumber())
	if err != nil {
		errCh <- err
	}

	addDevName := msclient.Option(
		func(d *kanbanpb.OutputRequest) error {
			d.NextDeviceName = "pluto"
			return nil
		},
	)
	addSendDataPath := msclient.Option(
		func(d *kanbanpb.OutputRequest) error {
			d.FileList = []string{"sendData/testData.png"}
			return nil
		},
	)
	ck := msclient.SetConnectionKey("slack")
	req, err := msclient.NewOutputData(ck, addDevName, addSendDataPath)
	if err != nil {
		errCh <- err
		return
	}
	err = c.OutputKanban(req)
	if err != nil {
		errCh <- err
		return
	}

loop:
	for {
		select {
		case err := <-errCh:
			log.Print(err)
			break loop
		case <-quiteCh:
			cancel()
		}
	}
}
