package app

import (
	"bitbucket.org/latonaio/aion-core/internal/kanban"
	"bitbucket.org/latonaio/aion-core/pkg/common"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	_struct "github.com/golang/protobuf/ptypes/struct"
)

type Session struct {
	io               kanban.Adapter
	microserviceName string
	cacheKanban      *kanbanpb.StatusKanban
	processNumber    int
	sendCh           chan *kanbanpb.Response
	dataPath         string
}

// create microservice session
func NewMicroserviceSessionWithRedis() *Session {
	return newSession(kanban.NewRedisAdapter())
}

// create microservice session
func NewMicroserviceSessionWithFile(dataPath string) *Session {
	return newSession(kanban.NewFileAdapter(dataPath))
}

// create struct of session with service broker
func newSession(io kanban.Adapter) *Session {
	sendCh := make(chan *kanbanpb.Response)
	return &Session{
		io:               io, // kanban io ( redis or directory )
		microserviceName: "",
		cacheKanban:      nil,
		processNumber:    0,
		sendCh:           sendCh,
	}
}

// start kanban watcher
func (s *Session) StartKanbanWatcher(ctx context.Context, p *kanbanpb.InitializeService) error {

	s.microserviceName = p.MicroserviceName
	s.processNumber = int(p.ProcessNumber)
	ch, err := s.io.WatchKanban(ctx, s.microserviceName, s.processNumber, kanban.StatusType_Before)
	if err != nil {
		return err
	}

	go func() {
		currentServiceData := &kanbanpb.ServiceData{
			Name: s.microserviceName,
		}
		for {
			select {
			case <-ctx.Done():
				close(s.sendCh)
				return
			case kanban, ok := <-ch:
				if !ok {
					close(s.sendCh)
					return
				}
				anyMsg, err := ptypes.MarshalAny(kanban)
				if err != nil {
					log.Printf("[kanban Watcher] cant unmarchal status kanban to any message")
				}
				s.cacheKanban = kanban
				s.cacheKanban.Services = append(s.cacheKanban.Services, currentServiceData)
				res := &kanbanpb.Response{
					MessageType: kanbanpb.ResponseType_RES_CACHE_KANBAN,
					Message:     anyMsg,
				}
				log.Printf("[KanbanWatcher] success to read kanban: (ms:%s, number:%d)",
					s.microserviceName, s.processNumber)
				s.sendCh <- res
			}
		}
	}()
	return nil
}

// set kanban from microservice
func (s *Session) SetKanban(p *kanbanpb.InitializeService) *kanbanpb.Response {
	res := &kanbanpb.Response{}
	res.MessageType = kanbanpb.ResponseType_RES_CACHE_KANBAN
	// get cache kanban
	s.cacheKanban = &kanbanpb.StatusKanban{
		StartAt:       common.GetIsoDatetime(),
		FinishAt:      "",
		Services:      []*kanbanpb.ServiceData{{Name: s.microserviceName}},
		ConnectionKey: "",
		ProcessNumber: p.ProcessNumber,
		DataPath:      common.GetMsDataPath(s.dataPath, p.MicroserviceName, int(p.ProcessNumber)),
		Metadata:      &_struct.Struct{},
	}
	s.microserviceName = p.MicroserviceName
	s.processNumber = int(p.ProcessNumber)

	msg, err := ptypes.MarshalAny(s.cacheKanban)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.Message = msg
	return res
}

// set next service yaml to output kanban
func (s *Session) OutputKanban(p *kanbanpb.OutputRequest) *kanbanpb.Response {
	res := &kanbanpb.Response{}
	res.MessageType = kanbanpb.ResponseType_RES_REQUEST_RESULT
	// check that already set microservice name
	if s.microserviceName == "" {
		res.Error = "input json is not read yet"
		return res
	}

	nextServiceData := &kanbanpb.ServiceData{
		Name:   "",
		Device: p.DeviceName,
	}
	// set metadata to after kanban
	afterKanban := *s.cacheKanban
	afterKanban.FinishAt = common.GetIsoDatetime()
	afterKanban.PriorSuccess = p.PriorSuccess
	afterKanban.FileList = p.FileList
	afterKanban.Metadata = p.Metadata
	afterKanban.Services = append(afterKanban.Services, nextServiceData)
	afterKanban.ProcessNumber = p.ProcessNumber
	afterKanban.ConnectionKey = p.ConnectionKey

	// write after kanban
	s.cacheKanban.StartAt = common.GetIsoDatetime()
	if err := s.io.WriteKanban(s.microserviceName, s.processNumber, &afterKanban, kanban.StatusType_After); err != nil {
		res.Error = fmt.Sprintf("cant write kanban: %v", err)
	}
	return res
}
