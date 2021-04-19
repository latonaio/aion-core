package app

import (
	"context"

	"bitbucket.org/latonaio/aion-core/internal/kanban"
	"bitbucket.org/latonaio/aion-core/pkg/common"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	_struct "github.com/golang/protobuf/ptypes/struct"
)

type Session struct {
	io               kanban.Adapter
	microserviceName string
	processNumber    int
	dataPath         string
}

// create microservice session
func NewMicroserviceSession(io kanban.Adapter, dataPath string, microservice *kanbanpb.InitializeService) *Session {
	session := newSession(io, microservice.MicroserviceName, int(microservice.ProcessNumber), dataPath)
	if microservice.InitType == kanbanpb.InitializeType_START_SERVICE_WITHOUT_KANBAN {
		session.setKanban()
	}
	return session
}

// create struct of session with service broker
func newSession(io kanban.Adapter, msName string, msNumber int, dataPath string) *Session {
	return &Session{
		io:               io, // kanban io ( redis or directory )
		microserviceName: msName,
		processNumber:    msNumber,
		dataPath:         dataPath,
	}
}

// start kanban watcher
func (s *Session) StartKanbanWatcher(ctx context.Context, sendCh chan<- *kanbanpb.StatusKanban) {

	defer func() {
		log.Printf("[KanbanWatcher] session closed (%s:%d)", s.microserviceName, s.processNumber)
	}()

	log.Printf("[KanbanWatcher] start session (%s:%d)", s.microserviceName, s.processNumber)
	s.io.WatchKanban(ctx, sendCh, s.microserviceName, s.processNumber, kanban.StatusType_Before, true)
}

// set kanban from microservice
func (s *Session) setKanban() error {
	// get cache kanban
	k := &kanbanpb.StatusKanban{
		StartAt:       common.GetIsoDatetime(),
		FinishAt:      "",
		ConnectionKey: "",
		ProcessNumber: int32(s.processNumber),
		DataPath:      common.GetMsDataPath(s.dataPath, s.microserviceName, s.processNumber),
		Metadata:      &_struct.Struct{},
	}

	if err := s.io.WriteKanban(s.microserviceName, s.processNumber, k, kanban.StatusType_Before); err != nil {
		log.Errorf("cannot create initial kanban: %v", err)
		return err
	}
	return nil
}
