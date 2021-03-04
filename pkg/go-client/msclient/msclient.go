// Copyright (c) 2019-2020 Latona. All rights reserved.

package msclient

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/encoding/protojson"
	"strings"
	"time"
)

// Option defines type of function that sets data to output request.
type Option func(*kanbanpb.OutputRequest) error

// NewOutputData constructs output request to kanban.
func NewOutputData(options ...Option) (*kanbanpb.OutputRequest, error) {
	d := &kanbanpb.OutputRequest{
		PriorSuccess:  true,
		DataPath:      "",
		ConnectionKey: "default",
		ProcessNumber: 1,
		FileList:      []string{},
		Metadata:      nil,
		DeviceName:    "",
	}
	for _, option := range options {
		if err := option(d); err != nil {
			return nil, err
		}
	}
	return d, nil
}

// SetResult returns option that microservice processing result.
func SetResult(r bool) Option {
	return func(d *kanbanpb.OutputRequest) error {
		d.PriorSuccess = r
		return nil
	}
}

// SetDataPath returns option that data path.
func SetDataPath(p string) Option {
	return func(d *kanbanpb.OutputRequest) error {
		d.DataPath = p
		return nil
	}
}

// SetConnectionKey returns option that connection key.
func SetConnectionKey(k string) Option {
	return func(d *kanbanpb.OutputRequest) error {
		d.ConnectionKey = k
		return nil
	}
}

// SetProcessNumber returns option that process number of microservice.
func SetProcessNumber(n int) Option {
	return func(d *kanbanpb.OutputRequest) error {
		d.ProcessNumber = int32(n)
		return nil
	}
}

// SetFileList returns option that file list.
func SetFileList(l []string) Option {
	return func(d *kanbanpb.OutputRequest) error {
		d.FileList = l
		return nil
	}
}

// SetMetadata converts metadata to format accept by kanban and returns option.
func SetMetadata(in map[string]interface{}) Option {
	return func(d *kanbanpb.OutputRequest) error {
		b, err := json.Marshal(in)
		if err != nil {
			return errors.Wrap(err, "cant marshal metadata to json binary: ")
		}
		s := &_struct.Struct{}
		if err := protojson.Unmarshal(b, s); err != nil {
			return errors.Wrap(err, "cant unmarshal metadata to protobuf struct: ")
		}
		d.Metadata = s
		return nil
	}
}

// SetDeviceName returns option that device name.
func SetDeviceName(n string) Option {
	return func(d *kanbanpb.OutputRequest) error {
		d.DeviceName = n
		return nil
	}
}

// MicroserviceClient declares communication function between kanban and microservice.
type MicroserviceClient interface {
	GetOneKanban(serviceName string, processNumber int) (*WrapKanban, error)
	GetKanbanCh(serviceName string, processNumber int) (chan *WrapKanban, error)
	SetKanban(serviceName string, processNumber int) (*WrapKanban, error)
	OutputKanban(data *kanbanpb.OutputRequest) error
	Close() error
	GetProcessNumber() int
}

// microserviceClient implements communication function between kanban and microservice.
type microserviceClient struct {
	stream       kanbanpb.Kanban_MicroserviceConnClient
	conn         *grpc.ClientConn
	env          Env
	sendCh       chan *kanbanpb.Request
	ackCh        chan string
	recvKanbanCh chan *WrapKanban
}

// Env has environment information.
type Env struct {
	KanbanAddr string
	MsNumber   int
	IsDocker   bool
}

// NewKanbanClient constructs kanban client object.
func NewKanbanClient(ctx context.Context) (MicroserviceClient, error) {
	// get env
	var env Env
	envconfig.Process("", &env)
	if env.KanbanAddr == "" {
		env.KanbanAddr = "localhost:11010"
	}
	if env.MsNumber < 1 {
		env.MsNumber = 1
	}

	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	// connect to send anything server
	conn, err := grpc.DialContext(ctx, env.KanbanAddr, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))

	if err != nil {
		conn.Close()
		return nil, errors.Wrap(err, fmt.Sprintf("cannot connect to status kanban server: %s", env.KanbanAddr))
	}
	log.Printf("success to connect status kanban server: %s", env.KanbanAddr)

	client := kanbanpb.NewKanbanClient(conn)
	stream, err := client.MicroserviceConn(ctx)
	if err != nil {
		conn.Close()
		return nil, errors.Wrap(err, fmt.Sprintf("cannot connect to status kanban server: %s", env.KanbanAddr))
	}

	// send function
	sendCh := make(chan *kanbanpb.Request, 1)
	go func() {
		for r := range sendCh {
			if r == nil {
				log.Printf("receive function is finished therefore send function is ended")
				break
			}
			if err := stream.Send(r); err != nil {
				log.Printf("cant send request to kanban server(message type: %s): %v", r.MessageType, err)
			} else {
				log.Printf("success to send request(message type: %s)", r.MessageType)
			}
		}
	}()

	// receive function
	ackCh := make(chan string, 5)
	recvKanbanCh := make(chan *WrapKanban, 5)
	go func() {
		for {
			m, err := stream.Recv()
			if err != nil {
				log.Printf("connection with kanban server is closed: %v", err)
				close(ackCh)
				close(recvKanbanCh)
				break
			}
			switch m.MessageType {
			case kanbanpb.ResponseType_RES_CACHE_KANBAN:
				if m.Error != "" {
					log.Printf("cant get kanban data: %s", m.Error)
					recvKanbanCh <- nil
					continue
				}
				var k kanbanpb.StatusKanban
				if err := ptypes.UnmarshalAny(m.Message, &k); err != nil {
					log.Printf("cant unmarshal any message to kanban message: %v", err)
					recvKanbanCh <- nil
					continue
				}
				recvKanbanCh <- &WrapKanban{k}
			case kanbanpb.ResponseType_RES_REQUEST_RESULT:
				if m.Error != "" {
					log.Printf("response is error: %s", m.Error)
				}
				ackCh <- m.Error
			}
		}
	}()

	c := &microserviceClient{
		stream:       stream,
		sendCh:       sendCh,
		ackCh:        ackCh,
		recvKanbanCh: recvKanbanCh,
		conn:         conn,
		env:          env,
	}
	return c, nil
}

func (k *microserviceClient) sendRequest(messageType kanbanpb.RequestType, body proto.Message) error {
	any, err := ptypes.MarshalAny(body)
	if err != nil {
		return errors.Wrap(err, "cant marshal protobuf message to any type")
	}
	m := &kanbanpb.Request{
		MessageType: messageType,
		Message:     any,
	}
	k.sendCh <- m
	return nil
}

func (k *microserviceClient) sendKanbanRequest(messageType kanbanpb.RequestType, serviceName string, processNumber int) error {
	m := &kanbanpb.InitializeService{
		MicroserviceName: serviceName,
		ProcessNumber:    int32(processNumber),
	}
	if err := k.sendRequest(messageType, m); err != nil {
		return err
	}
	return nil
}

// SetKanban sets service name and process number to kanban.
func (k *microserviceClient) SetKanban(serviceName string, processNumber int) (*WrapKanban, error) {
	if err := k.sendKanbanRequest(kanbanpb.RequestType_START_SERVICE_WITHOUT_KANBAN, serviceName, processNumber); err != nil {
		return nil, err
	}
	select {
	case <-time.After(time.Millisecond * 100):
		return nil, fmt.Errorf("timeout of waiting response by kanban server")
	case kan := <-k.recvKanbanCh:
		if kan == nil {
			return nil, fmt.Errorf("setting kanban is failed")
		}
		return kan, nil
	}
}

// GetOneKanban gets one kanban.
func (k *microserviceClient) GetOneKanban(serviceName string, processNumber int) (*WrapKanban, error) {
	if err := k.sendKanbanRequest(kanbanpb.RequestType_START_SERVICE, serviceName, processNumber); err != nil {
		return nil, err
	}
	select {
	case <-time.After(time.Millisecond * 100):
		return nil, fmt.Errorf("timeout of waiting response by kanban server")
	case kan := <-k.recvKanbanCh:
		if kan == nil {
			return nil, fmt.Errorf("getting kanban is failed")
		}
		return kan, nil
	}
}

// GetKanbanCh gets kanban channel.
func (k *microserviceClient) GetKanbanCh(serviceName string, processNumber int) (chan *WrapKanban, error) {
	if err := k.sendKanbanRequest(kanbanpb.RequestType_START_SERVICE, serviceName, processNumber); err != nil {
		return nil, err
	}
	return k.recvKanbanCh, nil
}

// OutputKanban outputs request to kanban.
func (k *microserviceClient) OutputKanban(data *kanbanpb.OutputRequest) error {
	if err := k.sendRequest(kanbanpb.RequestType_OUTPUT_AFTER_KANBAN, data); err != nil {
		return fmt.Errorf("output kanban is failed: %v", err)
	}
	select {
	case <-time.After(time.Second):
		return fmt.Errorf("cant get ack of output kanban request")
	case err := <-k.ackCh:
		if err != "" {
			return fmt.Errorf("failed to output kanban: %v", err)
		}
	}
	return nil
}

// Close closes the microservice client.
func (k *microserviceClient) Close() error {
	var errStr []string
	if k.stream != nil {
		if err := k.stream.CloseSend(); err != nil {
			errStr = append(errStr, err.Error())
		}
	}
	if k.conn != nil {
		if err := k.conn.Close(); err != nil {
			errStr = append(errStr, err.Error())
		}
	}
	if len(errStr) != 0 {
		return fmt.Errorf("cant close connection :\n%s", strings.Join(errStr, "\n "))
	}
	return nil
}

// GetProcessNumber gets process number.
func (k *microserviceClient) GetProcessNumber() int {
	return k.env.MsNumber
}
