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

	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
)

type Server struct {
	env *Env
}

// connect to redis server and start grpc server
func NewServer(env *Env) error {
	server := &Server{
		env: env,
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

func sendResponse(stream kanbanpb.Kanban_MicroserviceConnServer, res *kanbanpb.Response, session *Session) error {
	if err := stream.Send(res); err != nil {
		log.Printf(
			"grpc send error (name: %s, number: %d): %v",
			session.microserviceName,
			session.processNumber,
			err,
		)
		return err
	}
	log.Printf(
		"send message to microservice (name: %s, number: %d, message type: %s)",
		session.microserviceName,
		session.processNumber,
		res.MessageType,
	)
	return nil
}

// callback function when receive message from microservice
func (srv *Server) MicroserviceConn(stream kanbanpb.Kanban_MicroserviceConnServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer func() {
		log.Printf("connection closed")
		cancel()
	}()

	var session *Session
	// create redis pool when recieve gRPC call is no reasonable in terms of speed.
	// but we should do this becase must close connection to don't overflow block xread connection.
	redis := my_redis.NewRedisClient(srv.env.GetRedisAddr())
	session = NewMicroserviceSessionWithRedis(redis)

	session.dataPath = srv.env.GetDataDir()

	// receive kanban from redis and send client microservice
	recvCh := make(chan *kanbanpb.Response)
	go func() {
		defer close(recvCh)
		// receive kanban from client microservice and write kanban in redis
		for {
			in, err := stream.Recv()
			if err != nil {
				log.Printf(
					"receive stream is closed (name: %s, number: %d): %v",
					session.microserviceName,
					session.processNumber,
					err,
				)
				return
			}
			res, terminated, err := parseRequestMessage(ctx, session, in)
			if terminated {
				log.Printf("received terminate message: (%s:%d)", session.microserviceName, session.processNumber)
				return
			}
			if err != nil {
				log.Errorf(
					"cant parse request (name: %s, number %d):  %v",
					session.microserviceName,
					session.processNumber,
					err,
				)
				res.Error = err.Error()
			}
			// send write kanban result
			if res != nil {
				recvCh <- res
			}
		}
	}()

	for {
		select {
		case res, ok := <-session.sendCh:
			if !ok {
				return nil
			}
			if res.Error != "" {
				log.Printf(
					"grpc server error (name: %s, number: %d): %s",
					session.microserviceName,
					session.processNumber,
					res.Error,
				)
			}
			if err := sendResponse(stream, res, session); err != nil {
				return err
			}
		case res, ok := <-recvCh:
			if !ok {
				return nil
			}
			if res.Error != "" {
				log.Printf(
					"kanban parse error (name: %s, number: %d): %s",
					session.microserviceName,
					session.processNumber,
					res.Error,
				)
			}
			if err := sendResponse(stream, res, session); err != nil {
				return err
			}
		}
	}
}

func parseRequestMessage(ctx context.Context, session *Session, req *kanbanpb.Request) (*kanbanpb.Response, bool, error) {
	message := req.GetRequestMessage()
	switch t := message.(type) {
	case *kanbanpb.Request_InitMessage:
		return startService(ctx, session, t.InitMessage)
	case *kanbanpb.Request_Message:
		return outputAfterKanban(ctx, session, t.Message)
	}
	return nil, false, fmt.Errorf("message type is not defined")
}

func startService(ctx context.Context, session *Session, m *kanbanpb.InitializeService) (*kanbanpb.Response, bool, error) {
	var res *kanbanpb.Response = nil
	if m.InitType == kanbanpb.InitializeType_START_SERVICE_WITHOUT_KANBAN {
		res = session.SetKanban(m)
	}
	if err := session.StartKanbanWatcher(ctx, m); err != nil {
		log.Printf("cant start kanban watcher: %v", err)
		return nil, false, err
	}
	return res, false, nil
}

func outputAfterKanban(ctx context.Context, session *Session, m *kanbanpb.StatusKanban) (*kanbanpb.Response, bool, error) {
	terminated := false
	res, terminated := session.OutputKanban(m)
	return res, terminated, nil
}
