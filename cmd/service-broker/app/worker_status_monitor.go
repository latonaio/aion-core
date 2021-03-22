package app

import (
	"encoding/json"

	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
)

const redisKey = "aion-cluster-status"

type WorkerStatusMonitor struct {
	Rdisc    *my_redis.RedisClient
	RdisKey  string
	WorkerCh <-chan map[string]map[string]bool
}

type AionStatus map[string]bool

// MarshalBinary　redis hset する際にmarshalが必要なため
func (m AionStatus) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func NewWorkerStatusMonitor(nmch <-chan map[string]map[string]bool) *WorkerStatusMonitor {
	return &WorkerStatusMonitor{
		Rdisc:    my_redis.GetInstance(),
		RdisKey:  redisKey,
		WorkerCh: nmch,
	}
}

// Start workerのservice　statusの状態変更を受信しredisを更新
func (wm *WorkerStatusMonitor) Start() {
	//初期化
	_, err := wm.Rdisc.Delete(wm.RdisKey)
	if err != nil {
		log.Fatal(err)
	}

	for {
		// blocking
		meta := <-wm.WorkerCh

		nodeMap := map[string]interface{}{}
		for k, v := range meta {
			serviceMap := make(AionStatus)
			for k2, bl := range v {
				serviceMap[k2] = bl
			}
			nodeMap[k] = serviceMap
		}

		log.Debugf("[master]redis hset key: %v ,value: %+v \n", wm.RdisKey, nodeMap)
		// 状態記録
		_, err = wm.Rdisc.HSet(wm.RdisKey, nodeMap)
		if err != nil {
			log.Printf("[WorkerMonitor][start][redis_hset] failed cause:", err)
		}
	}
}
