package kanban

import (
	"context"

	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
)

// watcher and writer adapter
type Adapter interface {
	WriteKanban(msName string, msNumber int, kanban *kanbanpb.StatusKanban, statusType StatusType) error
	WatchKanban(ctx context.Context, msName string, msNumber int, statusType StatusType) (<-chan *kanbanpb.StatusKanban, error)
}

type StatusType int

const (
	StatusType_Before StatusType = iota
	StatusType_After
)
