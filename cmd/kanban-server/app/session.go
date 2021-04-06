package app

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"bitbucket.org/latonaio/aion-core/internal/kanban"
	"bitbucket.org/latonaio/aion-core/pkg/common"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"github.com/golang/protobuf/ptypes"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/protobuf/encoding/protojson"
)

type Session struct {
	sync.Mutex
	io               kanban.Adapter
	microserviceName string
	cacheKanban      *kanbanpb.StatusKanban
	processNumber    int
	sendCh           chan *kanbanpb.Response
	dataPath         string
	isActive         bool
}

// create microservice session
func NewMicroserviceSessionWithRedis(redis *my_redis.RedisClient) *Session {
	return newSession(kanban.NewRedisAdapter(redis))
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
		isActive:         false,
	}
}

func (s *Session) IsActive() bool {
	return s.isActive
}

func (s *Session) activate() {
	s.Lock()
	s.isActive = true
	s.Unlock()
}

func (s *Session) deactivate() {
	s.Lock()
	close(s.sendCh)
	s.isActive = false
	s.Unlock()
}

// start kanban watcher
func (s *Session) StartKanbanWatcher(ctx context.Context, p *kanbanpb.InitializeService) error {

	if s.IsActive() {
		return fmt.Errorf("session already activate")
	}
	s.activate()
	s.microserviceName = p.MicroserviceName
	s.processNumber = int(p.ProcessNumber)

	childCtx, cancel := context.WithCancel(ctx)
	ch := make(chan *kanbanpb.StatusKanban)
	go s.io.WatchKanban(childCtx, ch, s.microserviceName, s.processNumber, kanban.StatusType_Before, true)

	go func() {
		currentServiceData := &kanbanpb.ServiceData{
			Name: s.microserviceName,
		}
		for {
			select {
			case <-ctx.Done():
				s.deactivate()
				cancel()
				return
			case kanban, ok := <-ch:
				if !ok {
					s.deactivate()
					cancel()
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
	log.Printf("connect from microservice (%s:%d)", s.microserviceName, s.processNumber)

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
func (s *Session) OutputKanban(p *kanbanpb.OutputRequest) (*kanbanpb.Response, bool) {
	res := &kanbanpb.Response{}
	res.MessageType = kanbanpb.ResponseType_RES_REQUEST_RESULT
	// check that already set microservice name
	if s.microserviceName == "" {
		res.Error = "input json is not read yet"
		return res, false
	}

	// set metadata to after kanban
	afterKanban := s.cacheKanban
	afterKanban.FinishAt = common.GetIsoDatetime()
	afterKanban.PriorSuccess = p.PriorSuccess
	afterKanban.FileList = p.FileList
	afterKanban.Metadata = p.Metadata
	afterKanban.ProcessNumber = p.ProcessNumber
	afterKanban.ConnectionKey = p.ConnectionKey

	for _, v := range afterKanban.Services {
		v.Device = p.DeviceName
	}

	// write after kanban
	s.cacheKanban.StartAt = common.GetIsoDatetime()
	if err := s.io.WriteKanban(s.microserviceName, s.processNumber, afterKanban, kanban.StatusType_After); err != nil {
		res.Error = fmt.Sprintf("cant write kanban: %v", err)
		return res, false
	}

	var metadata map[string]interface{}
	b, err := protojson.Marshal(p.Metadata)
	if err != nil {
		res.Error = "cannot marshal metadata"
		return res, false
	}
	if err := json.Unmarshal(b, &metadata); err != nil {
		res.Error = "cannot unmarshal json"
		return res, false
	}

	if t, ok := metadata["type"].(string); ok && t == "terminate" {
		return res, true
	}

	return res, false
}
