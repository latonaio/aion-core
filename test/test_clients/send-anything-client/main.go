package main

import (
	"context"
	"log"
	"runtime"
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
	quiteCh := make(chan syscall.Signal, 1)

	ctx, cancel := context.WithCancel(context.Background())
	c, err := msclient.NewKanbanClient(ctx, msName, kanbanpb.InitializeType_START_SERVICE_WITHOUT_KANBAN)
	errStop(err)

	addDevName := msclient.Option(
		func(d *kanbanpb.StatusKanban) error {
			d.NextDeviceName = "pluto"
			return nil
		},
	)
	addSendDataPath := msclient.Option(
		func(d *kanbanpb.StatusKanban) error {
			if d.FileList == nil {
				d.FileList = make([]string, 0, 1)
			}
			d.FileList = append(d.FileList, "sendData/sendfile.txt")
			return nil
		},
	)
	ck := msclient.SetConnectionKey("slack")
	req, err := msclient.NewOutputData(ck, addDevName, addSendDataPath)
	errStop(err)

	err = c.OutputKanban(req)
	errStop(err)

	successStop()
	<-quiteCh
	cancel()
	return
}

func errStop(err error) {
	if err == nil {
		return
	}

	_, f, l, _ := runtime.Caller(1)
	log.Printf("ERROR in %s:%d", f, l)
	log.Printf("Process is alive and waiting. err -----")
	log.Printf("%v", err)
	for {
	}
}

func successStop() {
	log.Printf("Success. process is alive and waiting.")
}
