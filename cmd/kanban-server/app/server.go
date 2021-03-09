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

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
)

type Server struct {
	env     *Env
	isRedis bool
}

// connect to redis server and start grpc server
func NewServer(env *Env) error {
	server := &Server{
		env:     env,
		isRedis: true,
	}

	// start grpc server
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(env.GetServerPort()))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Minute,
		PermitWithoutStream: true,
	}

	grpcServer := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep))
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

// callback function when receive message from microservice
func (srv *Server) MicroserviceConn(stream kanbanpb.Kanban_MicroserviceConnServer) error {
	ctx := stream.Context()
	log.Printf("connect from microservice")

	var session *Session
	// create redis pool when recieve gRPC call is no reasonable in terms of speed.
	// but we should do this becase must close connection to don't overflow block xread connection.
	if err := my_redis.GetInstance().CreatePool(srv.env.GetRedisAddr()); err != nil {
		log.Printf("cant connect to redis, use directory mode: %v", err)
		session = NewMicroserviceSessionWithFile(srv.env.GetAionHome())
		srv.isRedis = false
	} else {
		session = NewMicroserviceSessionWithRedis()
	}

	session.dataPath = srv.env.GetDataDir()

	// get message from client
	// and then parse message type
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				log.Printf("receive stream is closed: %v", err)
				return
			}
			res, err := parseRequestMessage(ctx, session, in)
			if err != nil {
				res.Error = err.Error()
			}
			if res != nil {
				if err := stream.Send(res); err != nil {
					log.Printf("grpc send error: %v", err)
				}
			}
		}
	}()

	// loop of send channel to microservice
	go func() {
		for res := range session.sendCh {
			if res.Error != "" {
				log.Printf("grpc server error: %s", res.Error)
			}
			if err := stream.Send(res); err != nil {
				log.Printf("grpc send error: %v", err)
			}
			log.Printf("send message to microservice (name: %s, number: %d, message type: %s)",
				session.microserviceName, session.processNumber, res.MessageType)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("connection closed")
			return nil
		}
	}
}

func parseRequestMessage(ctx context.Context, session *Session, m *kanbanpb.Request) (*kanbanpb.Response, error) {
	switch m.MessageType {
	case kanbanpb.RequestType_START_SERVICE:
		p := &kanbanpb.InitializeService{}
		if err := ptypes.UnmarshalAny(m.Message, p); err != nil {
			return nil, fmt.Errorf("failer unmarshal message in set next service request: %v", err)
		}
		if err := session.StartKanbanWatcher(ctx, p); err != nil {
			log.Printf("cant start kanban watcher: %v", err)
			return nil, err
		}
		return nil, nil

	case kanbanpb.RequestType_START_SERVICE_WITHOUT_KANBAN:
		p := &kanbanpb.InitializeService{}
		if err := ptypes.UnmarshalAny(m.Message, p); err != nil {
			return nil, fmt.Errorf("failer unmarshal message in set next service request: %v", err)
		}
		res := session.SetKanban(p)
		if err := session.StartKanbanWatcher(ctx, p); err != nil {
			log.Printf("cant start kanban watcher: %v", err)
			return nil, err
		}
		return res, nil

	case kanbanpb.RequestType_OUTPUT_AFTER_KANBAN:
		p := &kanbanpb.OutputRequest{}
		if err := ptypes.UnmarshalAny(m.Message, p); err != nil {
			return nil, fmt.Errorf("failed unmarshal message in set next service request: %v", err)
		}
		res := session.OutputKanban(p)
		return res, nil

	default:
		return nil, fmt.Errorf("message type is not defined: %s", m.MessageType)
	}
}
