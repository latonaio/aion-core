// Copyright (c) 2019-2020 Latona. All rights reserved.
package app

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/devices"
	"bitbucket.org/latonaio/aion-core/internal/kanban"
	"bitbucket.org/latonaio/aion-core/internal/microservice"
	"bitbucket.org/latonaio/aion-core/pkg/k8s"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
)

const defaultPort = 11001

type controller struct {
	sync.Mutex
	aionHome            string
	microserviceList    map[string]*microservice.ScaleContainer
	deviceController    *devices.Controller
	watcher             *Watcher
	aionSetting         *config.AionSetting
	microservicesStatus map[string]bool
	updateTrigger       chan map[string]bool
	config              *Config
	k8sEnv              *k8s.K8sEnv
}

func (msc *controller) setMicroserviceList() error {
	msList := msc.aionSetting.GetMicroserviceList()
	var errList []string
	for msName, msData := range msList {
		// set aion path
		msData.Env["AION_HOME"] = msc.aionHome
		msData.Env["DEVICE_NAME"] = msc.aionSetting.GetDeviceName()
		msData.Env["DEBUG"] = msc.aionSetting.GetDebug()

		if err := msc.setMicroservice(msName, msData); err != nil {
			errList = append(errList, err.Error())
			continue
		}

		// workerの場合のみ状態を更新
		if msc.config.IsWorkerMode() {
			msc.SetServiceStatusDeactivate(msName)
		}
	}
	log.Debugf("set　microservice counter: %v", len(msc.microserviceList))
	if len(errList) != 0 {
		log.Printf("set setMicroservice failed cause: %+v \n", errList)
	}
	return nil
}

func (msc *controller) setMicroservice(msName string, msData *config.Microservice) error {
	sc, err := microservice.NewScaleContainer(msc.k8sEnv, msc.aionHome, msName, msData)
	if err != nil {
		return err
	}
	msc.Lock()
	msc.microserviceList[msName] = sc
	msc.Unlock()
	log.Debugf("setMicroservice %v:", msName)
	return nil
}

func (msc *controller) startStartupMicroservice(ctx context.Context) error {
	msList := msc.aionSetting.GetMicroserviceList()
	var errList []string

	for msName, msData := range msList {
		// TODO :insert periodic executiopn
		if msData.Startup {
			err := msc.startMicroservice(ctx, msName)
			if err != nil {
				errList = append(errList, err.Error())
				// skip
				continue
			}
			// workerの時のみ 状態更新
			if msc.config.IsWorkerMode() {
				msc.SetServiceStatusActive(msName)
			}
		}
	}

	if len(errList) != 0 {
		errStr := strings.Join(errList, "\n")
		return fmt.Errorf("cant startup microservice: \n%s", errStr)
	}
	log.Debugf("start upped service count: %v", len(msList))
	return nil
}

func (msc *controller) startMicroservice(ctx context.Context, name string) error {
	ms, ok := msc.microserviceList[name]
	if !ok {
		return fmt.Errorf("there is no microservice: %s", name)
	}

	for i := 1; i <= ms.GetScale(); i++ {
		if err := ms.StartMicroservice(i); err != nil {
			log.Errorf("cannot start microservice (name:%s, num:%d)", name, i)
		} else {
			go msc.watcher.WatchMicroservice(ctx, name, i)
			log.Printf("start microservice: (name:%s, num:%d)", name, i)
		}
	}
	log.Printf("start microservice: %s", name)

	return nil
}

func (msc *controller) startMicroserviceByNum(ctx context.Context, name string, mNum int) error {
	ms, ok := msc.microserviceList[name]
	if !ok {
		log.Errorf("there is no microservice: %s", name)
		return fmt.Errorf("there is no microservice: %s", name)
	}
	if err := ms.StartMicroservice(mNum); err != nil {
		log.Errorf("failed to start microservice: (name:%s, num:%d", name, mNum)
		return err
	}
	// set deploy status active
	if msc.config.IsWorkerMode() {
		msc.SetServiceStatusActive(name)
	}

	go msc.watcher.WatchMicroservice(ctx, name, mNum)
	log.Printf("start microservice from watcher (name:%s, num:%d)", name, mNum)
	return nil
}

func (msc *controller) StopAllMicroservice() {
	for name, _ := range msc.microserviceList {
		if err := msc.stopMicroservice(name); err != nil {
			log.Warnf("there is no microservice: %s", name)
		}
	}
}

func (msc *controller) stopMicroservice(name string) error {
	ms, ok := msc.microserviceList[name]
	if !ok {
		return fmt.Errorf("there is no microservice: %s", name)
	}
	for i := 1; i <= ms.GetScale(); i++ {
		if err := ms.StopMicroservice(i); err != nil {
			log.Debugf("%v", err)
			log.Warnf("cannot stop microservice (name:%s, num:%d)", name, i)
		}
	}
	log.Printf("stop all microservice from watcher (name:%s)", name)
	// deploy status
	if msc.config.IsWorkerMode() {
		msc.SetServiceStatusDeactivate(name)
	}

	return nil
}

func (msc *controller) WatchKanbanForMicroservice(ctx context.Context, aionCh <-chan *config.AionSetting, aionChForWatcher chan<- *config.AionSetting) {
	var cancel context.CancelFunc
	var childCtx context.Context
	startCh := msc.watcher.GetStartCh()
	log.Println("WatchKanbanForMicroservice watching aionCh")
	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}
			log.Printf("stop watch kanban for microservice")
			return
		case as := <-aionCh:
			if cancel != nil {
				cancel()
			}
			child, cncl := context.WithCancel(ctx)
			cancel = cncl
			childCtx = child
			msc.StopAllMicroservice()
			msc.Lock()
			msc.aionSetting = as
			msc.Unlock()
			log.Debugf("[worker] deploy start with %+v", as)
			aionChForWatcher <- as
			if err := msc.setMicroserviceList(); err != nil {
				log.Errorf("setMicroserviceList failed cause: %v \n", err)
			}
			log.Debugf("setMicroserviceList done")
			if err := msc.startStartupMicroservice(childCtx); err != nil {
				log.Errorf("startStartupMicroservice failed cause: %v \n", err)
			}
		case ms := <-startCh:
			if err := msc.startMicroserviceByNum(childCtx, ms.Name, ms.Number); err != nil {
				log.Errorf("startMicroserviceByNum failed cause: %v \n", err)
			}
		}
	}
}

func StartMicroservicesController(ctx context.Context, env *Config, aionCh <-chan *config.AionSetting, redis *my_redis.RedisClient) (*controller, error) {
	// kanban use redis or file
	var adapter kanban.Adapter
	adapter = kanban.NewRedisAdapter(redis)
	if err := redis.FlushAll(); err != nil {
		log.Errorf("cant initialized redis: %v", err)
	}

	dc, err := devices.NewDeviceController(ctx, env.IsDocker())
	if err != nil {
		return nil, err
	}
	// start to watch result about previous microservice
	k8sEnv := k8s.NewK8sEnv(env.GetDataDir(), env.GetRepositoryPrefix(), env.GetNamespace(), env.GetRegistrySecret())
	msc := &controller{
		aionHome:            env.GetAionHome(),
		microserviceList:    make(map[string]*microservice.ScaleContainer),
		deviceController:    dc,
		microservicesStatus: make(map[string]bool),
		watcher:             NewWatcher(dc, adapter),
		updateTrigger:       make(chan map[string]bool, 1),
		config:              env,
		k8sEnv:              k8sEnv,
	}

	aionChForWatcher := make(chan *config.AionSetting)

	// watch receive channel from send anything server
	go msc.watcher.WatchReceiveKanban(ctx, aionChForWatcher)
	// wait to start microservice by watcher
	go msc.WatchKanbanForMicroservice(ctx, aionCh, aionChForWatcher)
	return msc, nil
}

func (msc *controller) GetMicroServicesStatus() map[string]bool {
	return msc.microservicesStatus
}

func (msc *controller) SetServiceStatusActive(serviceName string) {
	msc.Lock()
	msc.microservicesStatus[serviceName] = true
	msc.sendUpdateStatusTrigger(msc.microservicesStatus)
	msc.Unlock()

}
func (msc *controller) SetServiceStatusDeactivate(serviceName string) {
	msc.Lock()
	msc.microservicesStatus[serviceName] = false
	msc.Unlock()
	msc.sendUpdateStatusTrigger(msc.microservicesStatus)
}

func (msc *controller) sendUpdateStatusTrigger(updated map[string]bool) {
	msc.updateTrigger <- updated
}

func (msc *controller) GetStatusUpdateTrigger() <-chan map[string]bool {
	return msc.updateTrigger
}
