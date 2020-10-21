package kanban

import "bitbucket.org/latonaio/aion-core/proto/kanbanpb"

// reader and writer adapter
type Adaptor interface {
	ReadKanban(msName string, msNumber int, statusType StatusType) (*kanbanpb.StatusKanban, error)
	WriteKanban(msName string, msNumber int, kanban *kanbanpb.StatusKanban, statusType StatusType) error
	WatchKanban(msName string, msNumber int, statusType StatusType) (chan *kanbanpb.StatusKanban, error)
}

type StatusType int

const (
	StatusType_Before StatusType = iota
	StatusType_After
)
