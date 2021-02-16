package app

/*
	grpc server in kanban server
		connection with: microservices
*/

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"strconv"
	"time"
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

	return grpcServer.Serve(listen)
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
		session = NewMicroserviceSessionWithFile(ctx, srv.env.GetAionHome())
		srv.isRedis = false
	} else {
		session = NewMicroserviceSessionWithRedis(ctx)
	}

	session.dataPath = srv.env.GetDataDir()

	// create receive channel
	recvCh := make(chan *kanbanpb.Request, 1)
	go func() {
		for {
			// wait stream by client
			in, err := stream.Recv()
			if err != nil {
				log.Printf("receive stream is closed: %v", err)
				close(recvCh)
				return
			}
			recvCh <- in
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

loop:
	// loop in receive ch and ctx Done
	for {
		select {
		case <-ctx.Done():
			break loop
		case m, ok := <-recvCh:
			if !ok {
				break loop
			}
			// call message parser
			go srv.parseRequestMessage(ctx, session, m)
		}
	}

	// my_redis.GetInstance().Close()
	return nil
}

func (srv *Server) parseRequestMessage(ctx context.Context, session *Session, m *kanbanpb.Request) {
	streamRes := &kanbanpb.Response{}
	switch m.MessageType {
	case kanbanpb.RequestType_START_SERVICE:
		// func1: notification of starting microservice and start to request kanban
		p := &kanbanpb.InitializeService{}
		if err := ptypes.UnmarshalAny(m.Message, p); err != nil {
			streamRes.Error = fmt.Sprintf("failed unmarshal message in Get cache kanban request: %v", err)
			break
		}
		session.ReadKanban(p, streamRes)
		if err := session.StartKanbanWatcher(ctx); err != nil {
			log.Printf("cant start kanban watcher: %v", err)
		}

	case kanbanpb.RequestType_START_SERVICE_WITHOUT_KANBAN:
		// func2: notification of starting first microservice.
		// message include service name and process number
		p := &kanbanpb.InitializeService{}
		if err := ptypes.UnmarshalAny(m.Message, p); err != nil {
			streamRes.Error = fmt.Sprintf("failed unmarshal message in set next service request: %v", err)
			break
		}
		session.SetKanban(p, streamRes)
		if err := session.StartKanbanWatcher(ctx); err != nil {
			log.Printf("cant start kanban watcher: %v", err)
		}
	case kanbanpb.RequestType_OUTPUT_AFTER_KANBAN:
		// func3: request to output kanban to redis or directory from microservice
		p := &kanbanpb.OutputRequest{}
		if err := ptypes.UnmarshalAny(m.Message, p); err != nil {
			streamRes.Error = fmt.Sprintf("failed unmarshal message in set next service request: %v", err)
			break
		}
		session.OutputKanban(p, streamRes)
	default:
		// err function
		streamRes.Error = fmt.Sprintf("message type is not defined: %s", m.MessageType)
		log.Printf(streamRes.Error)
	}
	// sent response to client
	session.sendCh <- streamRes
}
