package services

import (
	"context"
	"sync"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	pb "bitbucket.org/latonaio/aion-core/proto/projectpb"
)

type ProjectServer struct {
	sync.Mutex
	pb.UnimplementedProjectServer
	AionCh   chan<- *config.AionSetting
	IsDocker bool
	prj      *pb.AionSetting
}

func (p *ProjectServer) Apply(ctx context.Context, prj *pb.AionSetting) (*pb.Response, error) {
	log.Printf("received aionSetting")
	p.Lock()
	p.prj = prj
	p.Unlock()
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

func (p *ProjectServer) Delete(ctx context.Context, prj *pb.AionSetting) (*pb.Response, error) {
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
