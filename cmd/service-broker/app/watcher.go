// Copyright (c) 2019-2020 Latona. All rights reserved.

package app

import (
	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/devices"
	"bitbucket.org/latonaio/aion-core/internal/kanban"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"context"
	"fmt"
	"strconv"
	"sync"
)

const (
	serviceBrokerName = "service-broker"
)

type Watcher struct {
	sync.Mutex
	kanban.Adapter
	startCh          chan *Container
	stopCh           chan *Container
	deviceController *devices.Controller
	aionSetting      *config.AionSetting
}

func NewWatcher(dc *devices.Controller, io kanban.Adapter) *Watcher {
	return &Watcher{
		Adapter:          io,
		startCh:          NewContainerCh(),
		stopCh:           NewContainerCh(),
		deviceController: dc,
	}
}

func (w *Watcher) WatchReceiveKanban(aionCh <-chan *config.AionSetting) {
	deviceCh := w.deviceController.GetReceiveKanbanCh()
	for {
		select {
		case as := <-aionCh:
			w.Lock()
			w.aionSetting = as
			w.Unlock()
		case k := <-deviceCh:
			if err := w.sendToNextService(k.AfterKanban, k.NextService, int(k.NextNumber)); err != nil {
				log.Print(err)
			}
		}
	}
}

func (w *Watcher) WatchMicroservice(ctx context.Context, msName string, msNumber int) {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	kanbanCh, err := w.WatchKanban(childCtx, msName, msNumber, kanban.StatusType_After)
	if err != nil {
		log.Printf("cannot start watch microservice (name:%s, num:%d)", msName, msNumber)
		return
	}

	log.Printf("[watcher] start watch microservice : %s-%03d\n", msName, msNumber)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[watcher] stop watch microservice : %s-%03d\n", msName, msNumber)
			return
		case k := <-kanbanCh:
			nextServiceList, err := w.aionSetting.GetNextServiceList(msName, k.ConnectionKey)
			if err != nil {
				log.Printf("[watcher] %v, skipped", err)
				continue
			}
			for _, nextService := range nextServiceList {
				number := config.GetNextNumber(k.ProcessNumber, nextService.NumberPattern)
				nextDeviceName := k.Services[len(k.Services)-1].Device
				if nextDeviceName == "" {
					nextDeviceName = nextService.NextDevice
				}
				if device, ok := w.aionSetting.GetDeviceList()[nextDeviceName]; ok {
					// send to other device
					w.deviceController.SendFileToDevice(nextDeviceName, k, nextService.NextServiceName, number, device.Addr)
				} else {
					// send to local microservice
					k.Services[len(k.Services)-1].Device = w.aionSetting.GetDeviceName()
					if err := w.sendToNextService(k, nextService.NextServiceName, number); err != nil {
						log.Print(err)
					}
				}
			}
		}
	}
}

func (w *Watcher) sendToNextService(k *kanbanpb.StatusKanban, serviceName string, number int) error {
	if serviceName == serviceBrokerName {
		serviceName, number, err := w.terminateServiceParser(k)
		if err != nil {
			return fmt.Errorf("[watcher: terminate microservice] %v", err)
		}
		w.stopCh <- NewContainer(serviceName, number)
	} else {
		if err := w.WriteKanban(serviceName, number, k, kanban.StatusType_Before); err != nil {
			return fmt.Errorf("[watcher: start microservice] %v", err)
		}
		w.startCh <- NewContainer(serviceName, number)
	}
	return nil
}

func (w *Watcher) terminateServiceParser(k *kanbanpb.StatusKanban) (string, int, error) {
	if funcName, ok := k.Metadata.Fields["type"]; !ok || funcName.String() != "terminate" {
		return "", 0, fmt.Errorf("invalid function name (expect: terminate)")
	}

	serviceNameValue, ok := k.Metadata.Fields["name"]
	if !ok {
		return "", 0, fmt.Errorf("not set service name")
	}
	serviceName := serviceNameValue.String()

	numberValue, ok := k.Metadata.Fields["number"]
	if !ok {
		return serviceName, -1, nil
	} else {
		number, err := strconv.Atoi(numberValue.String())
		if err != nil {
			return "", 0, fmt.Errorf("invalid number value :%s", numberValue)
		}
		return serviceName, number, nil
	}
}

func (w *Watcher) GetStartCh() chan *Container {
	return w.startCh
}

func (w *Watcher) GetStopCh() chan *Container {
	return w.stopCh
}

func NewContainerCh() chan *Container {
	return make(chan *Container)
}

func NewContainer(name string, number int) *Container {
	return &Container{Name: name, Number: number}
}

type Container struct {
	Name   string
	Number int
}
