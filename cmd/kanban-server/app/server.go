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
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

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

	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}

	kasp := keepalive.ServerParameters{
		Time:    2 * time.Minute,
		Timeout: 1 * time.Minute,
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
	)
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
	ctx, cancel := context.WithCancel(stream.Context())
	defer func() {
		log.Printf("[ReceiveKanban] connection closed: %s", req.MicroserviceName)
		cancel()
	}()

	// create redis pool when recieve gRPC call is no reasonable in terms of speed.
	// but we should do this becase must close connection to don't overflow block xread connection.
	session := NewMicroserviceSession(srv.io, req)
	session.dataPath = srv.env.GetDataDir()

	recvCh := make(chan *kanbanpb.StatusKanban)
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		if err := session.StartKanbanWatcher(ctx, recvCh); err != nil {
			log.Errorf("[ReceiveKanban] failed to stop watch kanban: %v", err)
			errCh <- err
		}
	}()

	// receive kanban from redis and send client microservice
	log.Printf("[ReceiveKanban] startconnection: %s", req.MicroserviceName)
	for {
		select {
		case res, ok := <-recvCh:
			if !ok {
				log.Printf("[ReceiveKanban] recvCh closed")
				return nil
			}
			if err := stream.Send(res); err != nil {
				log.Errorf("[ReceiveKanban] failed to send")
				return err
			}
		case err := <-errCh:
			return err
		}
	}
}

func (srv *Server) SendKanban(ctx context.Context, req *kanbanpb.Request) (*kanbanpb.Response, error) {
	if err := srv.io.WriteKanban(
		req.MicroserviceName,
		int(req.Message.ProcessNumber),
		req.Message,
		kanban.StatusType_After,
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
