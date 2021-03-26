package kanban

import (
	"context"

	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
)

// watcher and writer adapter
type Adapter interface {
	WriteKanban(msName string, msNumber int, kanban *kanbanpb.StatusKanban, statusType StatusType) error
	WatchKanban(ctx context.Context, kanbanCh chan<- *kanbanpb.StatusKanban, msName string, msNumber int, statusType StatusType, deleteOldKanban bool)
}

type StatusType int

const (
	StatusType_Before StatusType = iota
	StatusType_After
)
