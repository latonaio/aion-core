// Copyright (c) 2019-2020 Latona. All rights reserved.

package app

import (
	"context"
	"fmt"
	// "reflect"

	"strings"

	"bitbucket.org/latonaio/aion-core/internal/kanban"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_mongo"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"google.golang.org/protobuf/types/known/structpb"
)

type Watcher struct {
	kanban.Adapter
}

func NewRequestRedisWatcher() *Watcher {
	return newWatcher(kanban.NewRedisAdapter())
}

func newWatcher(io kanban.Adapter) *Watcher {
	return &Watcher{
		Adapter: io,
	}
}

func (w *Watcher) WatchMicroservice(ctx context.Context, msName string, msNumber int, statusType kanban.StatusType) {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	kanbanCh := make(chan *kanbanpb.StatusKanban)
	go w.WatchKanban(childCtx, kanbanCh, msName, msNumber, statusType, false)

	streamKey := getStreamKey(msName, msNumber, statusType)
	for {
		select {
		case <-ctx.Done():
			return
		case k, ok := <-kanbanCh:
			if !ok {
				return
			}
			w.WriteKanbanMongo(ctx, k, streamKey)
		}
	}
}

func (w *Watcher) WriteKanbanMongo(ctx context.Context, kanban *kanbanpb.StatusKanban, streamKey string) error {
	val := map[string]interface{}{
		"streamKey":     streamKey,
		"startAt":       kanban.StartAt,
		"finishAt":      kanban.FinishAt,
		"services":      kanban.Services,
		"connectionKey": kanban.ConnectionKey,
		"processNumber": kanban.ProcessNumber,
		"priorSuccess":  kanban.PriorSuccess,
		"dataPath":      kanban.DataPath,
		"fileList":      kanban.FileList,
		"metadata":      unpackStruct(kanban.Metadata),
	}

	if err := my_mongo.GetInstance().InsertOne(ctx, val); err != nil {
		log.Fatalf("[FATAL] %v", err)
	}
	log.Printf("[write kanban] write to mongodb (streamkey: %s)", streamKey)
	return nil
}

func getStreamKey(msName string, number int, statusType kanban.StatusType) string {
	switch statusType {
	case kanban.StatusType_Before:
		return strings.Join([]string{
			"kanban", "before", msName, fmt.Sprintf("%03d", number)}, ":")
	case kanban.StatusType_After:
		return strings.Join([]string{
			"kanban", "after", msName, fmt.Sprintf("%03d", number)}, ":")
	default:
		return ""
	}
}

func unpackStruct(data interface{}) interface{} {
	switch d := data.(type) {
	case *structpb.Struct:
		_metadata := make(map[string]interface{})
		for k, x := range d.Fields {
			_metadata[k] = unpackStruct(x.GetKind())
		}
		return _metadata
	case *structpb.Value_StructValue:
		_metadata := make(map[string]interface{})
		for k, x := range d.StructValue.Fields {
			_metadata[k] = unpackStruct(x.GetKind())
		}
		return _metadata
	case *structpb.Value_ListValue:
		_metadata := make([]interface{}, len(d.ListValue.Values))
		for i, x := range d.ListValue.Values {
			_metadata[i] = unpackStruct(x.GetKind())
		}
		return _metadata
	case *structpb.Value_StringValue:
		return d.StringValue
	case *structpb.Value_NumberValue:
		return d.NumberValue
	case *structpb.Value_BoolValue:
		return d.BoolValue
	case *structpb.Value_NullValue:
		return d.NullValue
	default:
		return d
	}
}
