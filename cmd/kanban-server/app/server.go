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

func sendResponse(stream kanbanpb.Kanban_MicroserviceConnServer, res *kanbanpb.Response, session *Session) {
	if err := stream.Send(res); err != nil {
		log.Printf(
			"grpc send error (name: %s, number: %d): %v",
			session.microserviceName,
			session.processNumber,
			err,
		)
	}
	log.Printf(
		"send message to microservice (name: %s, number: %d, message type: %s)",
		session.microserviceName,
		session.processNumber,
		res.MessageType,
	)
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
	if err := my_redis.GetInstance().CreatePool(srv.env.GetRedisAddr()); err != nil {
		log.Printf("cant connect to redis, use directory mode: %v", err)
		session = NewMicroserviceSessionWithFile(srv.env.GetAionHome())
	} else {
		session = NewMicroserviceSessionWithRedis()
	}

	session.dataPath = srv.env.GetDataDir()

	// receive kanban from redis and send client microservice
	recvCh := make(chan *kanbanpb.Response)
	go func() {
		for {
			select {
			case res, ok := <-session.sendCh:
				if !ok {
					return
				}
				if res.Error != "" {
					log.Printf(
						"grpc server error (name: %s, number: %d): %s",
						session.microserviceName,
						session.processNumber,
						res.Error,
					)
				}
				sendResponse(stream, res, session)
			case res, ok := <-recvCh:
				if !ok {
					return
				}
				if res.Error != "" {
					log.Printf(
						"kanban parse error (name: %s, number: %d): %s",
						session.microserviceName,
						session.processNumber,
						res.Error,
					)
				}
				sendResponse(stream, res, session)
			}

		}
	}()

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
			return nil
		}
		res, terminated, err := parseRequestMessage(ctx, session, in)
		if terminated {
			log.Printf("received terminate message: (%s:%d)", session.microserviceName, session.processNumber)
			return nil
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
}

func parseRequestMessage(ctx context.Context, session *Session, m *kanbanpb.Request) (*kanbanpb.Response, bool, error) {
	var terminated bool
	terminated = false

	switch m.MessageType {
	case kanbanpb.RequestType_START_SERVICE:
		p := &kanbanpb.InitializeService{}
		if err := ptypes.UnmarshalAny(m.Message, p); err != nil {
			return nil, terminated, fmt.Errorf("failer unmarshal message in set next service request: %v", err)
		}
		if err := session.StartKanbanWatcher(ctx, p); err != nil {
			log.Printf("cant start kanban watcher: %v", err)
			return nil, terminated, err
		}
		return nil, terminated, nil

	case kanbanpb.RequestType_START_SERVICE_WITHOUT_KANBAN:
		p := &kanbanpb.InitializeService{}
		if err := ptypes.UnmarshalAny(m.Message, p); err != nil {
			return nil, terminated, fmt.Errorf("failer unmarshal message in set next service request: %v", err)
		}
		res := session.SetKanban(p)
		if err := session.StartKanbanWatcher(ctx, p); err != nil {
			log.Printf("cant start kanban watcher: %v", err)
			return nil, terminated, err
		}
		return res, terminated, nil

	case kanbanpb.RequestType_OUTPUT_AFTER_KANBAN:
		p := &kanbanpb.OutputRequest{}
		if err := ptypes.UnmarshalAny(m.Message, p); err != nil {
			return nil, terminated, fmt.Errorf("failed unmarshal message in set next service request: %v", err)
		}
		res, terminated := session.OutputKanban(p)
		return res, terminated, nil

	default:
		return nil, terminated, fmt.Errorf("message type is not defined: %s", m.MessageType)
	}
}
