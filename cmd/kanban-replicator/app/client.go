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
	aionHome         string
	microserviceList map[string]*config.Microservice
	watcher          *Watcher
}

// connect to redis server and start grpc server
func NewClient(ctx context.Context, conf *Config) *Client {
	// create redis pool
	if err := my_redis.GetInstance().CreatePool(conf.getRedisAddr()); err != nil {
		log.Fatalf("cant connect to redis, exit kanban-replicator: %v", err)
	}

	// create mongo pool
	if err := my_mongo.GetInstance().CreatePool(ctx, conf.getMongoAddr(), conf.getKanbanDB(), conf.getKanbanCollection()); err != nil {
		log.Fatalf("cant connect to mongo, exit kanban-replicator: %v", err)
	}

	// assume always docker mode
	isDocker := true
	if err := config.GetInstance().LoadConfig(conf.getConfigPath(), isDocker); err != nil {
		log.Fatalf("cant open yaml file (path: %s): %v", conf.getConfigPath(), err)
	}

	return &Client{
		aionHome:         conf.getAionHome(),
		microserviceList: config.GetInstance().GetMicroserviceList(),
		watcher:          NewRequestRedisWatcher(),
	}
}

func (c *Client) StartWatchKanban(ctx context.Context, microserviceList map[string]*config.Microservice) {
	for msName, msData := range microserviceList {
		for i := 1; i <= msData.Scale; i++ {
			if !msData.WithoutKanban {
				if err := c.watcher.WatchMicroservice(ctx, msName, i, kanban.StatusType_After); err != nil {
					log.Printf("[ERR] %v", err)
				}
				if err := c.watcher.WatchMicroservice(ctx, msName, i, kanban.StatusType_Before); err != nil {
					log.Printf("[ERR] %v", err)
				}
			}
		}
	}
}

func (c *Client) GetWatcher() *Watcher {
	return c.watcher
}

func (c *Client) GetMicroServiceList() map[string]*config.Microservice {
	return c.microserviceList
}
