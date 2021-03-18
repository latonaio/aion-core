// Copyright (c) 2019-2020 Latona. All rights reserved.
package app

import (
	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/devices"
	"bitbucket.org/latonaio/aion-core/internal/kanban"
	"bitbucket.org/latonaio/aion-core/internal/microservice"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/my_redis"
	"context"
	"fmt"
	"strings"
	"sync"
)

const defaultPort = 11001

type controller struct {
	sync.Mutex
	aionHome         string
	microserviceList map[string]*microservice.ScaleContainer
	deviceController *devices.Controller
	watcher          *Watcher
	aionSetting      *config.AionSetting
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
	}
	return nil
}

func (msc *controller) setMicroservice(msName string, msData *config.Microservice) error {
	sc, err := microservice.NewScaleContainer(msc.aionHome, msName, msData)
	if err != nil {
		return err
	}
	msc.Lock()
	defer msc.Unlock()
	msc.microserviceList[msName] = sc
	return nil
}

func (msc *controller) startStartupMicroservice(ctx context.Context) error {
	msList := msc.aionSetting.GetMicroserviceList()
	var errList []string
	if len(errList) != 0 {
		errStr := strings.Join(errList, "\n")
		return fmt.Errorf(
			"detect invalid microservice configuration: \n%s", errStr)
	}

	for msName, msData := range msList {
		// TODO :insert periodic executiopn
		if msData.Startup {
			if err := msc.startMicroservice(ctx, msName); err != nil {
				errList = append(errList, err.Error())
			}
		}
	}

	if len(errList) != 0 {
		errStr := strings.Join(errList, "\n")
		return fmt.Errorf(
			"cant startup microservice: \n%s", errStr)
	}

	return nil
}

func (msc *controller) startMicroservice(ctx context.Context, name string) error {
	ms, ok := msc.microserviceList[name]
	if !ok {
		return fmt.Errorf("there is no microservice: %s", name)
	}
	for i := 1; i <= ms.GetScale(); i++ {
		if err := ms.StartMicroservice(i); err != nil {
			log.Printf("cannot start microservice (name:%s, num:%d)", name, i)
		}
		go msc.watcher.WatchMicroservice(ctx, name, i)
	}
	log.Printf("start microservice: %s", name)
	return nil
}

func (msc *controller) startMicroserviceByNum(ctx context.Context, name string, mNum int) error {
	ms, ok := msc.microserviceList[name]
	if !ok {
		log.Printf("there is no microservice: %s", name)
		return fmt.Errorf("there is no microservice: %s", name)
	}
	if err := ms.StartMicroservice(mNum); err != nil {
		log.Printf("failed to start microservice: (name:%s, num:%d", name, mNum)
		return err
	}
	go msc.watcher.WatchMicroservice(ctx, name, mNum)
	log.Printf("start microservice from watcher (name:%s, num:%d)", name, mNum)
	return nil
}

func (msc *controller) StopAllMicroservice() {
	for name, _ := range msc.microserviceList {
		if err := msc.stopMicroservice(name); err != nil {
			fmt.Errorf("there is no microservice: %s", name)
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
			log.Printf("cannot stop microservice (name:%s, num:%d)", name, i)
		}
	}
	log.Printf("stop all microservice from watcher (name:%s)", name)
	return nil
}

func (msc *controller) stopMicroserviceByNum(name string, mNum int) error {
	ms, ok := msc.microserviceList[name]
	if !ok {
		return fmt.Errorf("there is no microservice: %s", name)
	}
	if err := ms.StopMicroservice(mNum); err != nil {
		return err
	}
	log.Printf("stop microservice from watcher (name:%s, num:%d)", name, mNum)
	return nil
}

func (msc *controller) WatchKanbanForMicroservice(ctx context.Context, aionCh <-chan *config.AionSetting, aionChForWatcher chan<- *config.AionSetting) {
	var cancel context.CancelFunc
	var childCtx context.Context
	startCh := msc.watcher.GetStartCh()
	stopCh := msc.watcher.GetStopCh()
	for {
		select {
		case <-ctx.Done():
			return
		case as := <-aionCh:
			if cancel != nil {
				cancel()
			}
			child, cncl := context.WithCancel(ctx)
			cancel = cncl
			childCtx = child
			msc.Lock()
			msc.StopAllMicroservice()
			msc.aionSetting = as
			msc.Unlock()
			aionChForWatcher <- as
			if err := msc.setMicroserviceList(); err != nil {
				log.Printf("%v", err)
			}
			msc.startStartupMicroservice(childCtx)
		case ms := <-startCh:
			msc.startMicroserviceByNum(childCtx, ms.Name, ms.Number)
		case ms := <-stopCh:
			if ms.Number == -1 {
				msc.stopMicroservice(ms.Name)
			} else {
				msc.stopMicroserviceByNum(ms.Name, ms.Number)
			}
		}
	}
}

func StartMicroservicesController(ctx context.Context, env *Config, aionCh <-chan *config.AionSetting) (*controller, error) {
	dc, err := devices.NewDeviceController(ctx, env.IsDocker())
	if err != nil {
		return nil, err
	}

	var adapter kanban.Adapter
	if err := my_redis.GetInstance().CreatePool(env.GetRedisAddr()); err != nil {
		log.Printf("cant connect to redis, use directory mode: %v", err)
		adapter = kanban.NewFileAdapter(env.GetDataDir())
	} else {
		log.Printf("Use redis mode")
		adapter = kanban.NewRedisAdapter()
		if err := my_redis.GetInstance().FlushAll(); err != nil {
			log.Printf("cant initialized redis: %v", err)
		}
	}

	// start to watch result about previous microservice
	msc := &controller{
		aionHome:         env.GetAionHome(),
		microserviceList: make(map[string]*microservice.ScaleContainer),
		deviceController: dc,
	}

	aionChForWatcher := make(chan *config.AionSetting)
	msc.watcher = NewWatcher(dc, adapter)

	// watch receive channel from send anything server
	go msc.watcher.WatchReceiveKanban(aionChForWatcher)
	// wait to start microservice by watcher
	go msc.WatchKanbanForMicroservice(ctx, aionCh, aionChForWatcher)
	return msc, nil
}
