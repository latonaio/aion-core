package services

import (
	"context"
	"sync"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	devpb "bitbucket.org/latonaio/aion-core/proto/devicepb"
	prjpb "bitbucket.org/latonaio/aion-core/proto/projectpb"
	srvpb "bitbucket.org/latonaio/aion-core/proto/servicepb"
)

type MasterServer struct {
	sync.Mutex
	prjpb.UnimplementedProjectServer
	AionCh   chan<- *config.AionSetting
	IsDocker bool
	prj      *prjpb.AionSetting
}

func filter(f func(ms *srvpb.Microservice) bool, list map[string]*srvpb.Microservice) map[string]*srvpb.Microservice {
	filteredList := map[string]*srvpb.Microservice{}
	for i, v := range list {
		if f(v) {
			filteredList[i] = v
		}
	}
	return filteredList
}

func (m *MasterServer) Apply(ctx context.Context, prj *prjpb.AionSetting) (*prjpb.Response, error) {
	log.Printf("received aionSetting")
	m.Lock()
	m.prj = prj
	m.Unlock()

	for i, _ := range m.prj.Devices {
		p := &prjpb.AionSetting{
			Devices:       map[string]*devpb.Device{},
			Microservices: map[string]*srvpb.Microservice{},
		}
		p.Devices = m.prj.Devices
		p.Microservices = filter(
			func(ms *srvpb.Microservice) bool {
				return ms.TargetNode == i
			},
			m.prj.Microservices,
		)
		Apply("aion-service-broker-device:30644", p)
		Apply("aion-kanban-replicator--device:30655", p)
	}

	return &prjpb.Response{
		Message: "Success",
		Code:    prjpb.ResponseCode_OK,
	}, nil
}
