package kanban

import (
	"context"
	"fmt"
	"strings"

	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
)

// watcher and writer adapter
type Adapter interface {
	WriteKanban(streamKey string, kanban *kanbanpb.StatusKanban) error
	WatchKanban(ctx context.Context, kanbanCh chan<- *AdaptorKanban, streamKey string, deleteOldKanban bool)
	DeleteKanban(streamKey string, id string) error
}

type AdaptorKanban struct {
	ID     string
	Kanban *kanbanpb.StatusKanban
}

type StatusType int

const (
	StatusType_Before StatusType = iota
	StatusType_After
	StatusType_Static
)

func GetStreamKeyByStatusType(msName string, msNumber int, statusType StatusType) string {
	switch statusType {
	case StatusType_Before:
		return GetBeforeStreamKey(msName, msNumber)
	case StatusType_After:
		return GetAfterStreamKey(msName, msNumber)
	default:
		return ""
	}
}

func GetBeforeStreamKey(msName string, number int) string {
	return strings.Join([]string{
		"kanban", "before", msName, fmt.Sprintf("%03d", number)}, ":")
}

func GetAfterStreamKey(msName string, number int) string {
	return strings.Join([]string{
		"kanban", "after", msName, fmt.Sprintf("%03d", number)}, ":")
}

func GetStaticStreamKey(topic string) string {
	return strings.Join([]string{"kanban", "static", topic}, ":")
}
