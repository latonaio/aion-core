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

const initPrevID = "0"

func NewRedisAdapter(redis *my_redis.RedisClient) Adapter {
	return &RedisAdaptor{
		redis: redis,
	}
}

type RedisAdaptor struct {
	redis *my_redis.RedisClient
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

func (a *RedisAdaptor) WriteKanban(msName string, msNumber int, kanban *kanbanpb.StatusKanban, statusType StatusType) error {
	streamKey := getStreamKeyByStatusType(msName, msNumber, statusType)

	m := jsonpb.Marshaler{}
	kanbanJson, err := m.MarshalToString(kanban)
	if err != nil {
		return fmt.Errorf("[write kanban] cant marshal kanban, %v", err)
	}
	val := map[string]interface{}{"kanban": kanbanJson}
	if err := a.redis.XAdd(streamKey, val); err != nil {
		return fmt.Errorf("[write kanban] cant write kanban to redis, %v", err)
	}
	log.Printf("[write kanban] write to queue (streamkey: %s)", streamKey)
	return nil
}

func (a *RedisAdaptor) WatchKanban(ctx context.Context, kanbanCh chan<- *kanbanpb.StatusKanban, msName string, msNumber int, statusType StatusType, deleteOldKanban bool) {
	defer func() {
		log.Printf("[watch kanban] stop watch kanban %s:%d", msName, msNumber)
		close(kanbanCh)
	}()

	prevID := initPrevID
	streamKey := getStreamKeyByStatusType(msName, msNumber, statusType)
	ch := make(chan *kanbanpb.StatusKanban)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				log.Printf("[watch kanban] redis context closed")
				return
			default:
				hash, nextID, err := a.redis.XReadOne([]string{streamKey}, []string{prevID}, 1, 0)
				if err != nil {
					log.Errorf("[watch kanban] blocking in watching kanban is exit (streamKey :%s) %v", streamKey, err)
					return
				}
				if deleteOldKanban {
					log.Debugf("[watch kanban] remove already read kanban: (%s:%s)", streamKey, prevID)
					if err := a.redis.XDel(streamKey, []string{prevID}); err != nil {
						log.Errorf("[watch kanban] cannot delete kanban: (%s:%s)", streamKey, prevID)
					}
				}
				prevID = nextID
				k, err := unmarshalKanban(hash)
				if err != nil {
					log.Errorf("[watch kanban] %v (streamKey: %s)", err, streamKey)
					continue
				}
				log.Printf("[watch kanban] read by queue (streamKey: %s)", streamKey)
				ch <- k
			}
		}
	}()

	log.Printf("[watch kanban] start watch kanban %s:%d", msName, msNumber)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[watch kanban] context closed")
			return
		case k, ok := <-ch:
			if !ok {
				log.Printf("[watch kanban] redis channel closed")
				return
			}
			kanbanCh <- k
		}
	}
}
