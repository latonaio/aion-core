package app

/*
	grpc server in kanban server
		connection with: microservices
*/

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"google.golang.org/grpc"

	"bitbucket.org/latonaio/aion-core/internal/kanban"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
)

type Server struct {
	env *Env
	io  kanban.Adapter
}

// connect to redis server and start grpc server
func NewServer(env *Env) error {
	server := &Server{
		env: env,
		io:  kanban.NewRedisAdapter(my_redis.NewRedisClient(env.GetRedisAddr())),
	}

	// start grpc server
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(env.GetServerPort()))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	kanbanpb.RegisterKanbanServer(grpcServer, server)
	log.Printf("Start Status kanban server:%d", env.GetServerPort())

	errChan := make(chan error)
	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	select {
	case err := <-errChan:
		return err
	case <-quit:
		grpcServer.GracefulStop()
		return nil
	}
}

func (srv *Server) ReceiveKanban(req *kanbanpb.InitializeService, stream kanbanpb.Kanban_ReceiveKanbanServer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		log.Printf("[ReceiveKanban] connection closed: %s", req.MicroserviceName)
		cancel()
	}()

	session := NewMicroserviceSession(srv.io, srv.env.GetDataDir(), req)

	recvCh := make(chan *kanban.AdaptorKanban)
	go session.StartKanbanWatcher(ctx, recvCh)

	// receive kanban from redis and send client microservice
	log.Printf("[ReceiveKanban] startconnection: %s", req.MicroserviceName)
	for res := range recvCh {
		log.Printf("res.Kanban value: %v", res.Kanban)

		if err := stream.Send(res.Kanban); err != nil {
			log.Errorf("[ReceiveKanban] failed to send: %v", err)
			return err
		}
	}
	return nil
}

func (srv *Server) ReceiveStaticKanban(req *kanbanpb.Topic, stream kanbanpb.Kanban_ReceiveStaticKanbanServer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		log.Printf("[ReceiveStaticKanban] connection closed: %s", req.Name)
		cancel()
	}()

	recvCh := make(chan *kanban.AdaptorKanban)
	go srv.io.WatchKanban(ctx, recvCh, kanban.GetStaticStreamKey(req.Name), true)

	// receive kanban from redis and send client microservice
	log.Printf("[ReceiveStaticKanban] startconnection: %s", req.Name)
	for res := range recvCh {
		if err := stream.Send(&kanbanpb.StaticKanban{Id: res.ID, StatusKanban: res.Kanban}); err != nil {
			log.Errorf("[ReceiveStaticKanban] failed to send: %v", err)
			return err
		}
	}
	return nil
}

func (srv *Server) SendKanban(ctx context.Context, req *kanbanpb.Request) (*kanbanpb.Response, error) {
	if err := srv.io.WriteKanban(
		kanban.GetStreamKeyByStatusType(req.MicroserviceName, int(req.Message.ProcessNumber), kanban.StatusType_After),
		req.Message,
	); err != nil {
		log.Errorf("[SendKanban] failed to write kanban")
		return &kanbanpb.Response{
			Status: kanbanpb.ResponseStatus_FAILED,
			Error:  fmt.Sprintf("cannot write kanban: %v", err),
		}, err
	}

	return &kanbanpb.Response{
		Status: kanbanpb.ResponseStatus_SUCCESS,
		Error:  "",
	}, nil
}

func (srv *Server) SendStaticKanban(ctx context.Context, req *kanbanpb.StaticRequest) (*kanbanpb.Response, error) {
	if err := srv.io.WriteKanban(
		kanban.GetStaticStreamKey(req.Topic.Name),
		req.Message,
	); err != nil {
		log.Errorf("[SendStaticKanban] failed to write kanban")
		return &kanbanpb.Response{
			Status: kanbanpb.ResponseStatus_FAILED,
			Error:  fmt.Sprintf("cannot write kanban: %v", err),
		}, err
	}

	return &kanbanpb.Response{
		Status: kanbanpb.ResponseStatus_SUCCESS,
		Error:  "",
	}, nil
}

func (srv *Server) DeleteStaticKanban(ctx context.Context, req *kanbanpb.DeleteStaticRequest) (*kanbanpb.Response, error) {
	if err := srv.io.DeleteKanban(
		kanban.GetStaticStreamKey(req.Topic.Name),
		req.Id,
	); err != nil {
		log.Errorf("[DeleteStaticKanban] failed to delete kanban")
		return &kanbanpb.Response{
			Status: kanbanpb.ResponseStatus_FAILED,
			Error:  fmt.Sprintf("cannot delete kanban: %v", err),
		}, err
	}

	return &kanbanpb.Response{
		Status: kanbanpb.ResponseStatus_SUCCESS,
		Error:  "",
	}, nil
}
