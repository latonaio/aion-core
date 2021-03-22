// Copyright (c) 2019-2020 Latona. All rights reserved.
package microservice

import (
	"fmt"
	"strconv"

	"bitbucket.org/latonaio/aion-core/pkg/log"

	"bitbucket.org/latonaio/aion-core/config"
)

type Container interface {
	StartProcess() error
	StopAllProcess() error
}

type ContainerStatus struct {
	Container
	NumOfUpState int
}

type ScaleContainer struct {
	name          string
	containerList map[int]*ContainerStatus
}

func NewScaleContainer(aionHome string, msName string, msData *config.Microservice) (*ScaleContainer, error) {
	containerList := make(map[int]*ContainerStatus)
	var err error
	for i := 1; i <= int(msData.Scale); i++ {
		// Set ms number
		msData.Env["MS_NUMBER"] = strconv.Itoa(i)
		var ms Container
		log.Debugf("NewScaleContainer docker mode: %v", msData.Docker)
		if msData.Docker {
			msData.Env["IS_DOCKER"] = "true"
			ms = NewContainerMicroservice(msName, msData, i)
		} else {
			ms, err = NewDirectoryMicroservice(aionHome, msName, msData, i)
			if err != nil {
				return nil, err
			}
		}

		containerList[i] = &ContainerStatus{
			Container:    ms,
			NumOfUpState: 0,
		}
	}

	return &ScaleContainer{
		name:          msName,
		containerList: containerList,
	}, nil
}

func (sc *ScaleContainer) GetScale() int {
	return len(sc.containerList)
}

func (sc *ScaleContainer) StartMicroservice(mNum int) error {
	if len(sc.containerList) < mNum {
		return fmt.Errorf(
			"microservice number is over of scale "+
				"(name: %s, scale: %d, request: %d)",
			sc.name, len(sc.containerList), mNum)
	}
	if _, ok := sc.containerList[mNum]; !ok {
		return fmt.Errorf("microservice does not exists (name: %s, number:%d)", sc.name, mNum)
	}
	if sc.containerList[mNum].NumOfUpState > 0 {
		return fmt.Errorf(
			"microservice already started, multiple service is not allowed (name: %s, scale: %d, request: %d)",
			sc.name, len(sc.containerList), mNum)
	}
	if err := sc.containerList[mNum].StartProcess(); err != nil {
		return err
	}
	sc.containerList[mNum].NumOfUpState += 1
	return nil
}

func (sc *ScaleContainer) StopMicroservice(mNum int) error {
	if len(sc.containerList) < mNum {
		return fmt.Errorf(
			"microservice number is over of scale (name: %s, scale: %d, request: %d)",
			sc.name, len(sc.containerList), mNum)
	}
	if _, ok := sc.containerList[mNum]; !ok {
		return fmt.Errorf("microservice does not exists (name: %s, number:%d)", sc.name, mNum)
	}
	if sc.containerList[mNum].NumOfUpState == 0 {
		return fmt.Errorf(
			"microservice is already finished (name: %s, scale: %d, request: %d)",
			sc.name, len(sc.containerList), mNum)
	}
	if err := sc.containerList[mNum].StopAllProcess(); err != nil {
		return err
	}
	sc.containerList[mNum].NumOfUpState = 0
	return nil
}
