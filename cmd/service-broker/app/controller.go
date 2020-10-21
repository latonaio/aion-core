// Copyright (c) 2019-2020 Latona. All rights reserved.
package app

import (
	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/devices"
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

func (msc *controller) startMicroserviceByNum(name string, mNum int) error {
	ms, ok := msc.microserviceList[name]
	if !ok {
		return fmt.Errorf("there is no microservice: %s", name)
	}
	if err := ms.StartMicroservice(mNum); err != nil {
		return err
	}
	return nil
}

func (msc *controller) startMicroserviceByName(name string) error {
	ms, ok := msc.microserviceList[name]
	if !ok {
		return fmt.Errorf("there is no microservice: %s", name)
	}
	if err := ms.StartAllMicroservice(); err != nil {
		return err
	}
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
	return nil
}

func (msc *controller) stopMicroserviceByName(name string) error {
	ms, ok := msc.microserviceList[name]
	if !ok {
		return fmt.Errorf("there is no microservice: %s", name)
	}
	if err := ms.StopAllMicroservice(); err != nil {
		return err
	}
	return nil
}

func (msc *controller) stopMicroserviceByNameForInit(name string) error {
	ms, ok := msc.microserviceList[name]
	if !ok {
		return fmt.Errorf("there is no microservice: %s", name)
	}
	if err := ms.StopAllMicroserviceForInit(); err != nil {
		return err
	}
	return nil
}

func (msc *controller) deleteMicroservice(msName string) error {
	_, ok := msc.microserviceList[msName]
	if !ok {
		return fmt.Errorf("%s is already deleted", msName)
	}
	delete(msc.microserviceList, msName)
	return nil
}

func (msc *controller) setMicroserviceList() error {
	msList := config.GetInstance().GetMicroserviceList()
	var errList []string
	for msName, msData := range msList {
		// set aion path
		msData.Env["AION_HOME"] = msc.aionHome
		msData.Env["DEVICE_NAME"] = config.GetInstance().GetDeviceName()

		if err := msc.setMicroservice(msName, msData); err != nil {
			errList = append(errList, err.Error())
			continue
		}
		for i := 1; i <= msData.Scale; i++ {
			if !msData.WithoutKanban {
				if err := msc.watcher.WatchMicroservice(msName, i); err != nil {
					errList = append(errList, err.Error())
				}
			}
		}
	}
	if len(errList) != 0 {
		errStr := strings.Join(errList, "\n")
		return fmt.Errorf(
			"detect invalid microservice configuration: \n%s", errStr)
	}

	for msName, msData := range msList {
		// TODO :insert periodic executiopn
		if msData.Startup {
			if err := msc.startMicroserviceByName(msName); err != nil {
				errList = append(errList, err.Error())
			}
		} else {
			// terminate micro service in case of service-broker sudden terminate
			if err := msc.stopMicroserviceByNameForInit(msName); err != nil {
				log.Printf("[terminate] cant stop microservice :%v", err)
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

func (msc *controller) StopAllMicroservice() {
	for msName, _ := range msc.microserviceList {
		if err := msc.stopMicroserviceByName(msName); err != nil {
			log.Printf("[terminate] cant stop microservice :%v", err)
		}
	}
}

func (msc *controller) StopAllMicroserviceForInit() {
	for msName, _ := range msc.microserviceList {
		if err := msc.stopMicroserviceByName(msName); err != nil {
			log.Printf("[terminate] cant stop microservice :%v", err)
		}
	}
}

func (msc *controller) StartMicroserviceFromKanban() {
	for ms := range msc.watcher.GetReadyToStartCh().GetCh() {
		if err := msc.startMicroserviceByNum(ms.Name, ms.Number); err != nil {
			log.Println(err.Error())
			continue
		}
		log.Printf("start microservice from watcher (name:%s, num:%d)",
			ms.Name, ms.Number)
	}
}

func (msc *controller) StopMicroserviceFromKanban() {
	for ms := range msc.watcher.GetReadyToTerminateCh().GetCh() {
		if ms.Number == -1 {
			if err := msc.stopMicroserviceByName(ms.Name); err != nil {
				log.Println(err.Error())
				continue
			}
			log.Printf("stop all microservice from watcher (name:%s)",
				ms.Name, ms.Number)
		} else {
			if err := msc.stopMicroserviceByNum(ms.Name, ms.Number); err != nil {
				log.Println(err.Error())
				continue
			}
			log.Printf("stop microservice from watcher (name:%s, num:%d)",
				ms.Name, ms.Number)
		}
	}
}

func StartMicroservicesController(ctx context.Context, env *Config) (*controller, error) {
	dc, err := devices.NewDeviceController(ctx, env.IsDocker())
	if err != nil {
		return nil, err
	}

	// start to watch result about previous microservice
	msc := &controller{
		aionHome:         env.GetAionHome(),
		microserviceList: make(map[string]*microservice.ScaleContainer),
		deviceController: dc,
	}

	if err := my_redis.GetInstance().CreatePool(env.GetRedisAddr()); err != nil {
		log.Printf("cant connect to redis, use directory mode: %v", err)
		msc.watcher = NewRequestFileWatcher(dc, env.GetDataDir())
	} else {
		log.Printf("Use redis mode")
		msc.watcher = NewRequestRedisWatcher(dc)
		if err := my_redis.GetInstance().FlushAll(); err != nil {
			log.Printf("cant initialized redis: %v", err)
		}
	}

	// start microservice
	if err := msc.setMicroserviceList(); err != nil {
		return nil, err
	}

	// watch receive channel from send anything server
	go msc.watcher.WatchReceiveKanban()
	// wait to start microservice by watcher
	go msc.StartMicroserviceFromKanban()
	go msc.StopMicroserviceFromKanban()
	return msc, nil
}
