package services

import (
	"context"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	pb "bitbucket.org/latonaio/aion-core/proto/projectpb"
)

type ProjectServer struct {
	pb.UnimplementedProjectServer
	AionCh   chan<- *config.AionSetting
	IsDocker bool
	prj      *pb.AionSetting
}

func (p *ProjectServer) Apply(ctx context.Context, prj *pb.AionSetting) (*pb.Response, error) {
	log.Printf("received aionSetting")
	aionSetting, err := config.LoadConfigFromGRPC(prj, p.IsDocker)
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
