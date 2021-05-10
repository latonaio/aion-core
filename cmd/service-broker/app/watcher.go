// Copyright (c) 2019-2020 Latona. All rights reserved.

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/devices"
	"bitbucket.org/latonaio/aion-core/internal/kanban"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	serviceBrokerName = "service-broker"
)

type Watcher struct {
	sync.Mutex
	kanban.Adapter

	startCh chan *Container
	// send anything grpc server
	deviceController *devices.Controller
	aionSetting      *config.AionSetting
}

func NewWatcher(dc *devices.Controller, io kanban.Adapter) *Watcher {
	return &Watcher{
		Adapter:          io,
		startCh:          NewContainerCh(),
		deviceController: dc,
	}
}

func (w *Watcher) WatchReceiveKanban(ctx context.Context, aionCh <-chan *config.AionSetting) {
	deviceCh := w.deviceController.GetReceiveKanbanCh()
	for {
		select {
		case <-ctx.Done():
			log.Printf("[watcher] stop watch receive kanban")
			return
		case as := <-aionCh:
			w.Lock()
			w.aionSetting = as
			w.Unlock()
		case k, ok := <-deviceCh:
			if !ok {
				return
			}
			if err := w.WriteKanban(kanban.GetStreamKeyByStatusType(k.NextService, int(k.NextNumber), kanban.StatusType_Before), k.AfterKanban); err != nil {
				log.Errorf("[watcher: start microservice] %v", err)
			}
			w.startCh <- NewContainer(k.NextService, int(k.NextNumber))
		}
	}
}

func (w *Watcher) WatchMicroservice(ctx context.Context, msName string, msNumber int) {
	childCtx, cancel := context.WithCancel(ctx)
	defer func() {
		log.Printf("[watcher] stop watch microservice : %s-%03d\n", msName, msNumber)
		cancel()
	}()
	kanbanCh := make(chan *kanban.AdaptorKanban)
	go w.WatchKanban(childCtx, kanbanCh, kanban.GetStreamKeyByStatusType(msName, msNumber, kanban.StatusType_After), false)

	log.Printf("[watcher] start watch microservice : %s-%03d\n", msName, msNumber)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[watcher] stop watch microservice : %s-%03d\n", msName, msNumber)
			return
		case ak, ok := <-kanbanCh:
			if !ok {
				log.Warnf("[watcher] watch kanban closed")
				return
			}
			k := ak.Kanban
			if k.ConnectionKey == serviceBrokerName {
				serviceName, _, err := w.terminateServiceParser(k)
				if err != nil {
					log.Errorf("[watcher: terminate microservice] %v", err)
					continue
				}
				msList := w.aionSetting.GetMicroserviceList()
				msData, ok := msList[serviceName]
				if !ok {
					log.Errorf("[watcher: terminate microservice] not found such as service: %s", serviceName)
					continue
				}
				if !msData.Always {
					return
				}
			}
			nextServiceList, err := w.aionSetting.GetNextServiceList(msName, k.ConnectionKey)
			if err != nil {
				log.Warnf("[watcher] %v, skipped", err)
				continue
			}
			for _, nextService := range nextServiceList {
				number := config.GetNextNumber(k.ProcessNumber, nextService.NumberPattern)
				nextDeviceName := k.NextDeviceName
				if nextDeviceName == "" {
					nextDeviceName = nextService.NextDevice
				}
				if device, ok := w.aionSetting.GetDeviceList()[nextDeviceName]; ok {
					// send to other device
					w.deviceController.SendFileToDevice(nextDeviceName, k, nextService.NextServiceName, number, device.Addr)
				} else {
					// send to local microservice
					k.NextDeviceName = w.aionSetting.GetDeviceName()
					if err := w.WriteKanban(kanban.GetStreamKeyByStatusType(nextService.NextServiceName, number, kanban.StatusType_Before), k); err != nil {
						log.Errorf("[watcher: start microservice] %v", err)
					}
					w.startCh <- NewContainer(nextService.NextServiceName, number)
				}
			}
		}
	}
}

func (w *Watcher) terminateServiceParser(k *kanbanpb.StatusKanban) (string, int, error) {
	m := k.GetMetadata()
	var ret map[string]interface{}
	b, err := protojson.Marshal(m)
	if err != nil {
		return "", 0, fmt.Errorf("Failed to marshal grpc stream")
	}
	if err := json.Unmarshal(b, &ret); err != nil {
		return "", 0, fmt.Errorf("Failed to unmarshal json")
	}

	if funcName, ok := ret["type"].(string); !ok || funcName != "terminate" {
		log.Printf("ok %v funcName.String() %s", ok, funcName)
		return "", 0, fmt.Errorf("invalid function name (expect: terminate)")
	}

	serviceNameValue, ok := ret["name"].(string)
	if !ok {
		return "", 0, fmt.Errorf("not set service name")
	}

	number, ok := ret["number"].(int)
	if !ok {
		return serviceNameValue, -1, nil
	}
	return serviceNameValue, number, nil
}

func (w *Watcher) GetStartCh() chan *Container {
	return w.startCh
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
