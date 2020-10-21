package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/latonaio/aion-core/cmd/kanban-replicator/app"
	"bitbucket.org/latonaio/aion-core/pkg/log"
)

func main() {
	log.SetFormat("kanban-replicator")
	ctx := context.Background()

	client := app.NewClient(ctx, app.GetConfig())
	client.StartWatchKanban(ctx, client.GetMicroServiceList())

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
