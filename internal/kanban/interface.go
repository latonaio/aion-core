package kanban

import (
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"context"
)

// reader and writer adapter
type Adapter interface {
	ReadKanban(msName string, msNumber int, statusType StatusType) (*kanbanpb.StatusKanban, error)
	WriteKanban(msName string, msNumber int, kanban *kanbanpb.StatusKanban, statusType StatusType) error
	WatchKanban(ctx context.Context, msName string, msNumber int, statusType StatusType) (chan *kanbanpb.StatusKanban, error)
}

type StatusType int

const (
	StatusType_Before StatusType = iota
	StatusType_After
)
