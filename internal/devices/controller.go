// Copyright (c) 2019-2020 Latona. All rights reserved.
package devices

import (
	"context"
	"time"

	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"google.golang.org/grpc"
)

var addr string = "aion-sendanything:10000"

// TODO: change file sender module to adapter pattern
// for change to mqtt, etc...

type Controller struct {
	addr   string
	recvCh chan *kanbanpb.SendKanban
	sendCh chan *kanbanpb.SendKanban
}

func NewDeviceController(ctx context.Context) (*Controller, error) {

	c := &Controller{
		addr:   addr,
		recvCh: make(chan *kanbanpb.SendKanban, 1),
		sendCh: make(chan *kanbanpb.SendKanban, 1),
	}

	// connection function
	go c.connectToServer(ctx)
	return c, nil
}

func (c *Controller) connectToServer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(c.recvCh)
			return
		default:
			// send anything server 対する dial
			conn, err := grpc.DialContext(ctx, c.addr, grpc.WithInsecure())
			if err != nil {
				log.Printf("cant connect to send anything server: %s", c.addr)
				time.Sleep(time.Second * 5)
				continue
			}

			client := kanbanpb.NewSendAnythingClient(conn)
			stream, err := client.ServiceBrokerConn(ctx)
			if err != nil {
				log.Printf("cant connect to send anything server: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}
			log.Printf("success to connecting send anything server")

			// send to send-anything server
			doneCh := make(chan struct{})
			go func() {
				for {
					select {
					case <-doneCh:
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
					close(doneCh)
					if err := conn.Close(); err != nil {
						log.Printf("connection close failed cause: %v", err)
					}

					break
				}
				c.recvCh <- data
			}
		}
	}
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
