package services

import (
	"context"
	"sync"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
	pb "bitbucket.org/latonaio/aion-core/proto/projectpb"
)

type projectServer struct {
	sync.Mutex
	pb.UnimplementedProjectServer
	AionCh   chan<- *config.AionSetting
	IsDocker bool
	prj      *pb.AionSetting
	redis    *my_redis.RedisClient
}

func NewProjectServer(aionCh chan<- *config.AionSetting, isDocker bool, redis *my_redis.RedisClient) *projectServer {
	return &projectServer{
		AionCh:   aionCh,
		IsDocker: isDocker,
		prj:      nil,
		redis:    redis,
	}
}

func (p *projectServer) Apply(ctx context.Context, receivedAionSetting *pb.AionSetting) (*pb.Response, error) {
	log.Debugf("[grpc][server][received] AionSetting: %+v", receivedAionSetting)
	p.Lock()
	p.prj = receivedAionSetting
	p.Unlock()
	aionSetting, err := config.LoadConfigFromGRPC(p.prj, p.IsDocker)
	if err != nil {
		return &pb.Response{
			Message: "Failed",
			Code:    pb.ResponseCode_Failed,
		}, nil
	}

	p.AionCh <- aionSetting
	p.prj = receivedAionSetting

	return &pb.Response{
		Message: "Success",
		Code:    pb.ResponseCode_OK,
	}, nil
}

func (p *projectServer) Delete(ctx context.Context, prj *pb.AionSetting) (*pb.Response, error) {
	log.Printf("received aionSetting")

	for i, _ := range prj.Microservices {
		if _, exist := p.prj.Microservices[i]; exist {
			delete(p.prj.Microservices, i)
		}
	}

	aionSetting, err := config.LoadConfigFromGRPC(p.prj, p.IsDocker)
	if err != nil {
		return &pb.Response{
			Message: "Failed",
			Code:    pb.ResponseCode_Failed,
		}, nil
	}
	p.AionCh <- aionSetting
	p.prj = prj

	return &pb.Response{
		Message: "Success",
		Code:    pb.ResponseCode_OK,
	}, nil
}

func (p *projectServer) Status(context.Context, *pb.Empty) (*pb.Services, error) {
	result, err := p.redis.HGet("aion-cluster-status")
	if err != nil {
		log.Printf("[WorkerMonitor][GetAllServicesStatus] failed cause: %v", err)
		return &pb.Services{
			Response: &pb.Response{
				Message: err.Error(),
				Code:    pb.ResponseCode_Failed,
			},
			Status: nil,
		}, err
	}

	if len(result) == 0 {
		result = map[string]string{}
	}

	return &pb.Services{
		Response: &pb.Response{
			Message: "",
			Code:    pb.ResponseCode_Failed,
		},
		Status: result,
	}, nil
}
