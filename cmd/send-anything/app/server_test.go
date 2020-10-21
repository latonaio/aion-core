// Copyright (c) 2019-2020 Latona. All rights reserved.

package app

import (
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"context"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/grpc"
	"testing"
	"time"
)

var (
	testEnv = &Env{
		AionHome:   "../../../test/",
		ServerPort: 50010,
	}
	testSendKanban = &kanbanpb.SendKanban{
		DeviceName:  "test",
		DeviceAddr:  "localhost",
		NextService: "test2",
		NextNumber:  1,
		AfterKanban: &kanbanpb.StatusKanban{
			Services:      []string{"test1"},
			ProcessNumber: 1,
			PriorSuccess:  true,
			DataPath:      "../../../test/Data/test1_1/",
			FileList:      []string{"file/output/test.txt"},
			Metadata:      &_struct.Struct{},
		},
	}
	addr = "localhost:50010"
)

func TestServer_SendToOtherDevices(t *testing.T) {
	ctx := context.TODO()
	// start server
	s := &Server{
		env:                   testEnv,
		sendToServiceBrokerCh: make(chan *kanbanpb.SendKanban, 1),
	}
	go s.createServerConnection()

	// start client
	c := &Server{
		env:                   testEnv,
		sendToServiceBrokerCh: make(chan *kanbanpb.SendKanban, 1),
	}
	err := c.sendToOtherDeviceClient(ctx, testSendKanban)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-s.sendToServiceBrokerCh:
	case <-time.After(time.Second):
		t.Errorf("cant get response")
	}
}

// send kanban from A to B (service broker A -> send anything A -> send anything A)
func TestServer_ServiceBrokerConn(t *testing.T) {
	ctx := context.TODO()
	// start B send anything
	s := &Server{
		env:                   testEnv,
		sendToServiceBrokerCh: make(chan *kanbanpb.SendKanban, 1),
	}
	go s.createServerConnection()

	// start A service broker client
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("cant connect to send anything server: %s", addr)
	}

	client := kanbanpb.NewSendAnythingClient(conn)
	stream, err := client.ServiceBrokerConn(ctx)
	if err := stream.Send(testSendKanban); err != nil {
		t.Fatal(err)
	}

	select {
	case <-s.sendToServiceBrokerCh:
	case <-time.After(time.Second):
		t.Errorf("cant get response")
	}
}

// send kanban from A to B (service broker A -> send anything A -> send anything A -> service broker A)
func TestServer_ServiceBrokerConn2(t *testing.T) {
	ctx := context.TODO()
	// start A send anything
	go NewServer(testEnv)

	// start A service broker client
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("cant connect to send anything server: %s", addr)
	}

	client := kanbanpb.NewSendAnythingClient(conn)
	stream, err := client.ServiceBrokerConn(ctx)
	if err := stream.Send(testSendKanban); err != nil {
		t.Fatal(err)
	}
	ch := make(chan *kanbanpb.SendKanban, 1)
	go func() {
		k, err := stream.Recv()
		if err != nil {
			t.Fatal(err)
		}
		ch <- k
	}()

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Errorf("cant get response")
	}
}
