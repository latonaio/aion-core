// Copyright (c) 2019-2020 Latona. All rights reserved.

package app

import (
	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/devices"
	"bitbucket.org/latonaio/aion-core/internal/kanban"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"fmt"
	"strconv"
)

const (
	serviceBrokerName = "service-broker"
)

type Watcher struct {
	kanban.Adaptor
	readyToStartCh     *requestCh
	readyToTerminateCh *requestCh
	deviceController   *devices.Controller
}

func NewRequestFileWatcher(dc *devices.Controller, aionDataPath string) *Watcher {
	return NewWatcher(dc, kanban.NewFileAdapter(aionDataPath))
}

func NewRequestRedisWatcher(dc *devices.Controller) *Watcher {
	return NewWatcher(dc, kanban.NewRedisAdapter())
}

func NewWatcher(dc *devices.Controller, io kanban.Adaptor) *Watcher {
	return &Watcher{
		Adaptor:            io,
		readyToStartCh:     NewRequestCh(),
		readyToTerminateCh: NewRequestCh(),
		deviceController:   dc,
	}
}

func (w *Watcher) WatchReceiveKanban() {
	for k := range w.deviceController.GetReceiveKanbanCh() {
		if err := w.sendToNextService(k.AfterKanban, k.NextService, int(k.NextNumber)); err != nil {
			log.Print(err)
		}
	}
}

func (w *Watcher) WatchMicroservice(msName string, msNumber int) error {
	kanbanCh, err := w.WatchKanban(msName, msNumber, kanban.StatusType_After)
	if err != nil {
		return err
	}
	go func() {
		for k := range kanbanCh {
			nextServiceList, err := config.GetInstance().GetNextServiceList(msName, k.ConnectionKey)
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
				// send to other device
				if _, ok := config.GetInstance().GetDeviceList()[nextDeviceName]; ok {
					if err := w.deviceController.SendFileToDevice(
						nextDeviceName, k, nextService.NextServiceName, number); err != nil {
						log.Print(err)
					}
					// send to local microservice
				} else {
					k.Services[len(k.Services)-1].Device = config.GetInstance().GetDeviceName()
					if err := w.sendToNextService(k, nextService.NextServiceName, number); err != nil {
						log.Print(err)
					}
				}
			}
		}
	}()
	return nil
}

func (w *Watcher) sendToNextService(k *kanbanpb.StatusKanban, serviceName string, number int) error {
	if serviceName == serviceBrokerName {
		serviceName, number, err := w.terminateServiceParser(k)
		if err != nil {
			return fmt.Errorf("[watcher: terminate microservice] %v", err)
		}
		w.readyToTerminateCh.SetToCh(serviceName, number)
	} else {
		if err := w.WriteKanban(serviceName, number, k, kanban.StatusType_Before); err != nil {
			return fmt.Errorf("[watcher: start microservice] %v", err)
		}
		w.readyToStartCh.SetToCh(serviceName, number)
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

func (w *Watcher) GetReadyToStartCh() *requestCh {
	return w.readyToStartCh
}

func (w *Watcher) GetReadyToTerminateCh() *requestCh {
	return w.readyToTerminateCh
}

// ready to start channel
type requestCh struct {
	startCh chan *ReadyToStartContainer
}

func NewRequestCh() *requestCh {
	return &requestCh{startCh: make(chan *ReadyToStartContainer)}
}

func (c *requestCh) SetToCh(serviceName string, number int) {
	ns := &ReadyToStartContainer{
		Name:   serviceName,
		Number: number,
	}
	c.startCh <- ns
}

func (c *requestCh) GetCh() chan *ReadyToStartContainer {
	return c.startCh
}

type ReadyToStartContainer struct {
	Name   string
	Number int
}
