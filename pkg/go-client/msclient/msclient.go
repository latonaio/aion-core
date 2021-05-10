// Copyright (c) 2019-2020 Latona. All rights reserved.

package msclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"

	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/encoding/protojson"
)

// Option defines type of function that sets data to output request.
type Option func(*kanbanpb.StatusKanban) error

// NewOutputData constructs output request to kanban.
func NewOutputData(options ...Option) (*kanbanpb.StatusKanban, error) {
	d := &kanbanpb.StatusKanban{
		PriorSuccess:   true,
		DataPath:       "",
		ConnectionKey:  "default",
		ProcessNumber:  1,
		FileList:       []string{},
		Metadata:       nil,
		NextDeviceName: "",
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
	return func(d *kanbanpb.StatusKanban) error {
		d.PriorSuccess = r
		return nil
	}
}

// SetDataPath returns option that data path.
func SetDataPath(p string) Option {
	return func(d *kanbanpb.StatusKanban) error {
		d.DataPath = p
		return nil
	}
}

// SetConnectionKey returns option that connection key.
func SetConnectionKey(k string) Option {
	return func(d *kanbanpb.StatusKanban) error {
		d.ConnectionKey = k
		return nil
	}
}

// SetProcessNumber returns option that process number of microservice.
func SetProcessNumber(n int) Option {
	return func(d *kanbanpb.StatusKanban) error {
		d.ProcessNumber = int32(n)
		return nil
	}
}

// SetFileList returns option that file list.
func SetFileList(l []string) Option {
	return func(d *kanbanpb.StatusKanban) error {
		d.FileList = l
		return nil
	}
}

// SetMetadata converts metadata to format accept by kanban and returns option.
func SetMetadata(in map[string]interface{}) Option {
	return func(d *kanbanpb.StatusKanban) error {
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
	return func(d *kanbanpb.StatusKanban) error {
		d.NextDeviceName = n
		return nil
	}
}

// MicroserviceClient declares communication function between kanban and microservice.
type MicroserviceClient interface {
	GetOneKanban() (*kanbanpb.StatusKanban, error)
	GetKanbanCh() chan *kanbanpb.StatusKanban
	GetStaticKanban(ctx context.Context, topic string, kanbanCh chan *kanbanpb.StaticKanban) error
	OutputKanban(data *kanbanpb.StatusKanban) error
	OutputStaticKanban(topic string, data *kanbanpb.StatusKanban) error
	DeleteStaticKanban(topic string, id string) error
	Close() error
	GetProcessNumber() int
}

// microserviceClient implements communication function between kanban and microservice.
type microserviceClient struct {
	recvClient   kanbanpb.KanbanClient
	recvConn     *grpc.ClientConn
	sendClient   kanbanpb.KanbanClient
	sendConn     *grpc.ClientConn
	env          Env
	ackCh        chan string
	recvKanbanCh chan *kanbanpb.StatusKanban
	serviceName  string
}

// Env has environment information.
type Env struct {
	KanbanAddr string `envconfig:"KANBAN_ADDR"`
	MsNumber   int
	IsDocker   bool
}

// NewKanbanClient constructs kanban client object.
func NewKanbanClient(ctx context.Context, serviceName string, initType kanbanpb.InitializeType) (MicroserviceClient, error) {
	// get env
	var env Env
	envconfig.Process("", &env)
	if env.KanbanAddr == "" {
		env.KanbanAddr = "aion-statuskanban:10000" //statuskanbanのenvoy
	}
	if env.MsNumber < 1 {
		env.MsNumber = 1
	}

	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	// connect to status-kanban server
	recvConn, err := grpc.DialContext(ctx, env.KanbanAddr, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("cannot connect to status kanban server: %s", env.KanbanAddr))
	}

	sendConn, err := grpc.DialContext(ctx, env.KanbanAddr, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		recvConn.Close()
		return nil, errors.Wrap(err, fmt.Sprintf("cannot connect to status kanban server: %s", env.KanbanAddr))
	}

	log.Printf("success to connect status kanban server: %s", env.KanbanAddr)

	recvClient := kanbanpb.NewKanbanClient(recvConn)
	sendClient := kanbanpb.NewKanbanClient(sendConn)
	initMsg := &kanbanpb.InitializeService{
		InitType:         initType,
		MicroserviceName: serviceName,
		ProcessNumber:    int32(env.MsNumber),
	}
	stream, err := recvClient.ReceiveKanban(ctx, initMsg)
	if err != nil {
		recvConn.Close()
		return nil, errors.Wrap(err, fmt.Sprintf("cannot connect to status kanban server: %s", env.KanbanAddr))
	}

	recvKanbanCh := make(chan *kanbanpb.StatusKanban, 5)
	go func() {
		for {
			m, err := stream.Recv()
			if err != nil {
				log.Printf("connection with kanban server is closed: %v", err)
				close(recvKanbanCh)
				break
			}
			recvKanbanCh <- m
		}
	}()

	c := &microserviceClient{
		recvClient:   recvClient,
		recvConn:     recvConn,
		sendClient:   sendClient,
		sendConn:     sendConn,
		recvKanbanCh: recvKanbanCh,
		env:          env,
		serviceName:  serviceName,
	}
	return c, nil
}

func (k *microserviceClient) sendRequest(req *kanbanpb.Request) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()
	_, err := k.sendClient.SendKanban(ctx, req)
	if err != nil {
		log.Errorf("cannot send kanban:, %v", err)
		return err
	}
	return nil
}

// GetOneKanban gets one kanban.
func (k *microserviceClient) GetOneKanban() (*kanbanpb.StatusKanban, error) {
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
func (k *microserviceClient) GetKanbanCh() chan *kanbanpb.StatusKanban {
	return k.recvKanbanCh
}

func (k *microserviceClient) GetStaticKanban(ctx context.Context, topic string, kanbanCh chan *kanbanpb.StaticKanban) error {
	defer close(kanbanCh)
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	recvConn, err := grpc.DialContext(ctx, k.env.KanbanAddr, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cannot connect to status kanban server: %s", k.env.KanbanAddr))
	}
	defer recvConn.Close()
	recvClient := kanbanpb.NewKanbanClient(recvConn)
	initMsg := &kanbanpb.Topic{Name: topic}
	stream, err := recvClient.ReceiveStaticKanban(ctx, initMsg)
	if err != nil {
		recvConn.Close()
		return errors.Wrap(err, fmt.Sprintf("cannot connect to status kanban server: %s", k.env.KanbanAddr))
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			m, err := stream.Recv()
			if err != nil {
				log.Printf("connection with kanban server is closed: %v", err)
				return nil
			}
			kanbanCh <- m
		}
	}
}

// OutputKanban outputs request to kanban.
func (k *microserviceClient) OutputKanban(data *kanbanpb.StatusKanban) error {

	req := &kanbanpb.Request{
		MicroserviceName: k.serviceName,
		Message:          data,
	}
	if err := k.sendRequest(req); err != nil {
		return fmt.Errorf("output kanban is failed: %v", err)
	}
	return nil
}

func (k *microserviceClient) OutputStaticKanban(topic string, data *kanbanpb.StatusKanban) error {

	req := &kanbanpb.StaticRequest{
		Topic:   &kanbanpb.Topic{Name: topic},
		Message: data,
	}
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()
	_, err := k.sendClient.SendStaticKanban(ctx, req)
	if err != nil {
		log.Errorf("cannot send static kanban:, %v", err)
		return err
	}
	return nil
}

func (k *microserviceClient) DeleteStaticKanban(topic string, id string) error {

	req := &kanbanpb.DeleteStaticRequest{
		Topic: &kanbanpb.Topic{Name: topic},
		Id:    id,
	}
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()
	_, err := k.sendClient.DeleteStaticKanban(ctx, req)
	if err != nil {
		log.Errorf("cannot delete kanban:, %v", err)
		return err
	}
	return nil
}

func (k *microserviceClient) sendTerminateKanban() error {
	metadata := map[string]string{}
	metadata["type"] = "terminate"
	metadata["number"] = strconv.Itoa(k.env.MsNumber)
	metadata["name"] = k.serviceName
	b, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("cant marshal metadata to json binary: %v", err)
	}

	s := &_struct.Struct{}
	if err := protojson.Unmarshal(b, s); err != nil {
		return fmt.Errorf("cant unmarshal metadata to protobuf struct: %v", err)
	}

	d := &kanbanpb.StatusKanban{
		PriorSuccess:   true,
		DataPath:       "",
		ConnectionKey:  "service-broker",
		ProcessNumber:  int32(k.env.MsNumber),
		FileList:       []string{},
		Metadata:       s,
		NextDeviceName: "",
	}

	if err := k.OutputKanban(d); err != nil {
		return fmt.Errorf("cant send terminate kanban: %v", err)
	}
	return nil
}

// Close closes the microservice client.
func (k *microserviceClient) Close() error {
	var errStr []string
	if k.sendClient != nil {
		if err := k.sendTerminateKanban(); err != nil {
			errStr = append(errStr, err.Error())
		}
	}
	if k.recvConn != nil {
		if err := k.recvConn.Close(); err != nil {
			errStr = append(errStr, err.Error())
		}
	}
	if k.sendConn != nil {
		if err := k.sendConn.Close(); err != nil {
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
