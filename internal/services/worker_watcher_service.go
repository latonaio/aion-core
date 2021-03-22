package services

import (
	"fmt"
	"sync"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/clusterpb"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

type WorkerWatcher struct {
	sync.Mutex
	ServicesStatus  map[string]map[string]bool
	ClientStream    map[string]clusterpb.Cluster_JoinMasterAionServer
	StatusMonitorCh chan<- map[string]map[string]bool
}

func NewWorkerWatcher(statusMonitorCh chan<- map[string]map[string]bool) *WorkerWatcher {
	return &WorkerWatcher{
		ServicesStatus:  make(map[string]map[string]bool),
		ClientStream:    make(map[string]clusterpb.Cluster_JoinMasterAionServer),
		StatusMonitorCh: statusMonitorCh,
	}
}

func (w *WorkerWatcher) JoinMasterAion(stream clusterpb.Cluster_JoinMasterAionServer) error {
	var nodeName string
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("[JoinMasterAion] no meta data")
	}
	nodeName = fmt.Sprintf("%s/%s", md.Get("NodeName")[0], md.Get("NodeIP")[0])
	log.Debugf("%v :初期化完了 \n", nodeName)
	w.Lock()
	w.ClientStream[nodeName] = stream
	w.Unlock()

	for {
		nm := &clusterpb.NodeMeta{}
		if err := stream.RecvMsg(nm); err != nil {
			log.Printf("[grpc] [error] JoinMasterAion RecvMsg failed cause:%v", err)
			break
		}

		// 更新
		w.Lock()
		w.ServicesStatus[nodeName] = nm.ServicesStatus
		w.Unlock()

		// statusMonitorに送信
		w.StatusMonitorCh <- w.ServicesStatus
		log.Debugf("[grpc][server] received service status change: %+v", w.ServicesStatus)
	}

	w.Lock()
	w.ServicesStatus[nodeName] = nil
	w.Unlock()
	return errors.New("grpc stream closed")
}

func (w *WorkerWatcher) SendAionSettingToWorker(aionSettingCh <-chan *config.AionSetting) {
	for {
		aionSetting := <-aionSettingCh
		for k, v := range aionSetting.Aion.Devices {
			key := fmt.Sprintf("%v/%v", k, v.Addr)
			v, ok := w.ClientStream[key]
			if !ok {
				continue
			}
			if err := v.SendMsg(&clusterpb.Apply{AionSetting: aionSetting.Aion}); err != nil {
				log.Printf("apply worker aionSetting failed cause: %v \n", err)
			}
			log.Debugf("ApplyWorkerAion", aionSetting)
		}
	}

}
