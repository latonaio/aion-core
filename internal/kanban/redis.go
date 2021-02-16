package kanban

import (
	"context"
	"fmt"
	"strings"

	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"github.com/golang/protobuf/jsonpb"
)

func NewRedisAdapter() Adapter {
	return &RedisAdaptor{
		prevID: "0",
	}
}

type RedisAdaptor struct {
	prevID string
}

func getBeforeStreamKey(msName string, number int) string {
	return strings.Join([]string{
		"kanban", "before", msName, fmt.Sprintf("%03d", number)}, ":")
}

func getAfterStreamKey(msName string, number int) string {
	return strings.Join([]string{
		"kanban", "after", msName, fmt.Sprintf("%03d", number)}, ":")
}

func getStreamKeyByStatusType(msName string, msNumber int, statusType StatusType) string {
	switch statusType {
	case StatusType_Before:
		return getBeforeStreamKey(msName, msNumber)
	case StatusType_After:
		return getAfterStreamKey(msName, msNumber)
	default:
		return ""
	}
}

func unmarshalKanban(hash map[string]interface{}) (*kanbanpb.StatusKanban, error) {
	str, ok := hash["kanban"].(string)
	if !ok {
		return nil, fmt.Errorf("kanban is not string")
	}

	u := jsonpb.Unmarshaler{}
	k := &kanbanpb.StatusKanban{}
	if err := u.Unmarshal(strings.NewReader(str), k); err != nil {
		return nil, fmt.Errorf("cant unmarshal kanban yaml: %v", err)
	}

	return k, nil
}

func (a *RedisAdaptor) ReadKanban(msName string, msNumber int, statusType StatusType) (*kanbanpb.StatusKanban, error) {
	streamKey := getStreamKeyByStatusType(msName, msNumber, statusType)
	hash, newID, err := my_redis.GetInstance().XReadOne([]string{streamKey}, []string{a.prevID}, 1, -1)
	if err != nil {
		return nil, fmt.Errorf("[read kanban] cant get by redis(key: %s): %v", streamKey, err)
	}
	a.prevID = newID
	if statusType == StatusType_Before {
		if err := my_redis.GetInstance().XDel(streamKey, []string{newID}); err != nil {
			log.Printf("camt delete keys (streamKey: %s, id: %s)", streamKey, newID)
		}
	}
	k, err := unmarshalKanban(hash)
	if err != nil {
		return nil, fmt.Errorf("[read kanban] %v (name: %s, number: %d)", err, msName, msNumber)
	}
	log.Printf("[read kanban] detect kanban (streamKey: %s)", streamKey)
	return k, nil
}

func (a *RedisAdaptor) WriteKanban(msName string, msNumber int, kanban *kanbanpb.StatusKanban, statusType StatusType) error {
	streamKey := getStreamKeyByStatusType(msName, msNumber, statusType)

	m := jsonpb.Marshaler{}
	kanbanJson, err := m.MarshalToString(kanban)
	if err != nil {
		return fmt.Errorf("[write kanban] cant marshal kanban, %v", err)
	}
	val := map[string]interface{}{"kanban": kanbanJson}
	if err := my_redis.GetInstance().XAdd(streamKey, val); err != nil {
		return fmt.Errorf("[write kanban] cant write kanban to redis, %v", err)
	}
	log.Printf("[write kanban] write to queue (streamkey: %s)", streamKey)
	return nil
}

func (a *RedisAdaptor) WatchKanban(ctx context.Context, msName string, msNumber int, statusType StatusType) (chan *kanbanpb.StatusKanban, error) {
	streamKey := getStreamKeyByStatusType(msName, msNumber, statusType)
	ch := make(chan *kanbanpb.StatusKanban)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			default:
				hash, nextID, err := my_redis.GetInstance().XReadOne([]string{streamKey}, []string{a.prevID}, 1, 0)
				if err != nil {
					log.Printf(
						"[watch kanban] blocking in watching kanban is exit (streamKey :%s) %v", streamKey, err)
					return
				}
				a.prevID = nextID
				k, err := unmarshalKanban(hash)
				if err != nil {
					log.Printf("[watch kanban] %v (streamKey: %s)", err, streamKey)
					continue
				}
				log.Printf("[watch kanban] read by queue (streamKey: %s)", streamKey)
				ch <- k
			}
		}
	}()
	return ch, nil
}
