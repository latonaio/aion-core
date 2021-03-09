// Copyright (c) 2019-2020 Latona. All rights reserved.
package devices

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"context"
	"google.golang.org/grpc"
)

// TODO: change file sender module to adapter pattern
// for change to mqtt, etc...

type Controller struct {
	addr   string
	recvCh chan *kanbanpb.SendKanban
	sendCh chan *kanbanpb.SendKanban
}

func NewDeviceController(ctx context.Context, isDocker bool) (*Controller, error) {
	var addr string
	if isDocker {
		addr = "aion-sendanything:10000"
	} else {
		addr = "localhost:11011"
	}

	c := &Controller{
		addr:   addr,
		recvCh: make(chan *kanbanpb.SendKanban, 1),
		sendCh: make(chan *kanbanpb.SendKanban, 1),
	}

	// connection function
	go func() {
		if err := c.connectToServer(ctx); err != nil {
			log.Printf("connection error")
		}
	}()

	return c, nil
}

func (c *Controller) connectToServer(ctx context.Context) error {
	conn, err := grpc.DialContext(ctx, c.addr, grpc.WithInsecure())
	if err != nil {
		log.Printf("cant connect to send anything server: %s", c.addr)
		return err
	}

	client := kanbanpb.NewSendAnythingClient(conn)
	stream, err := client.ServiceBrokerConn(ctx)
	if err != nil {
		log.Printf("cant connect to send anything server: %v", err)
		return err
	}
	log.Printf("success to connecting send anything server")

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(c.recvCh)
				return
			case data := <-c.sendCh:
				if err := stream.Send(data); err != nil {
					log.Printf("failed to send kanban to send-anything server: %s", data.DeviceName)
				} else {
					log.Printf("success to send kanban to send-anything server: %s", data.DeviceName)
				}
			}
		}
	}()

	// receive from send-anything server
	for {
		data, err := stream.Recv()
		if err != nil {
			log.Printf("connection closed from send anything server: %v", err)
			conn.Close()
			break
		}
		c.recvCh <- data
	}
	return nil
}

func (c *Controller) SendFileToDevice(deviceName string, k *kanbanpb.StatusKanban, nextService string, number int, addr string) {
	data := &kanbanpb.SendKanban{
		DeviceName:  deviceName,
		DeviceAddr:  addr,
		NextService: nextService,
		NextNumber:  int32(number),
		AfterKanban: k,
	}
	c.sendCh <- data
	log.Printf("set to sending queue of send-anything : %s", nextService)
}

func (c *Controller) GetReceiveKanbanCh() chan *kanbanpb.SendKanban {
	return c.recvCh
}
