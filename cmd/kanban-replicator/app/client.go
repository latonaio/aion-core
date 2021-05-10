package app

import (
	"context"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/kanban"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_mongo"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
)

type Client struct {
	aionHome    string
	watcher     *Watcher
	aionSetting *config.AionSetting
}

// connect to redis server and start grpc server
func NewClient(ctx context.Context, conf *Config, redis *my_redis.RedisClient) *Client {
	// create redis pool
	// create mongo pool
	if err := my_mongo.GetInstance().CreatePool(ctx, conf.GetMongoAddr(), conf.GetKanbanDB(), conf.GetKanbanCollection()); err != nil {
		log.Fatalf("cant connect to mongo, exit kanban-replicator: %v", err)
	}

	return &Client{
		aionHome: conf.GetAionHome(),
		watcher:  NewRequestRedisWatcher(redis),
	}
}

func (c *Client) StartWatchKanban(ctx context.Context, aionCh <-chan *config.AionSetting) {

	var cancel context.CancelFunc

	for {
		select {
		case <-ctx.Done():
			return
		case ya := <-aionCh:
			if cancel != nil {
				cancel()
			}
			childCtx, cncl := context.WithCancel(ctx)
			cancel = cncl
			microserviceList := ya.GetMicroserviceList()
			for msName, msData := range microserviceList {
				for i := 1; i <= int(msData.Scale); i++ {
					if !msData.WithoutKanban {
						go c.watcher.WatchMicroservice(childCtx, msName, i, kanban.StatusType_After)
						go c.watcher.WatchMicroservice(childCtx, msName, i, kanban.StatusType_Before)
					}
				}
			}
		}
	}
}
