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

func unmarshalKanban(hash map[string]interface{}) (*kanbanpb.StatusKanban, error) {
	if hash == nil {
		log.Println("hash not found")
		return nil,nil
	}
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

func (a *RedisAdaptor) WriteKanban(streamKey string, kanban *kanbanpb.StatusKanban) error {

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

func (a *RedisAdaptor) WatchKanban(ctx context.Context, kanbanCh chan<- *AdaptorKanban, streamKey string, deleteOldKanban bool) {
	defer func() {
		log.Printf("[watch kanban] stop watch kanban :%s", streamKey)
		close(kanbanCh)
	}()

	prevID := initPrevID
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
			kanbanCh <- &AdaptorKanban{ID: prevID, Kanban: k}
		}
	}
}

func (a *RedisAdaptor) DeleteKanban(streamKey string, id string) error {
	if err := a.redis.XDel(streamKey, []string{id}); err != nil {
		log.Errorf("[delete kanban] cannot delete kanban: (%s:%s)", streamKey, id)
		return err
	}
	return nil
}
