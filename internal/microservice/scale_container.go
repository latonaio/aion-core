// Copyright (c) 2019-2020 Latona. All rights reserved.
package microservice

import (
	"bitbucket.org/latonaio/aion-core/config"
	"fmt"
	"strconv"
	"strings"
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
	data          *config.Microservice
}

func NewScaleContainer(aionHome string, msName string, msData *config.Microservice) (*ScaleContainer, error) {
	containerList := make(map[int]*ContainerStatus)
	var err error
	for i := 1; i <= msData.Scale; i++ {
		// Set ms number
		msData.Env["MS_NUMBER"] = strconv.Itoa(i)
		var ms Container
		if msData.Docker {
			msData.Env["IS_DOCKER"] = "true"
			ms = NewContainerMicroservice(msName, msData, i)
		} else {
			ms, err = NewDirectoryMicroservice(aionHome, msName, msData, i)
			if err != nil {
				return nil, err
			}
		}
		cs := &ContainerStatus{
			Container:    ms,
			NumOfUpState: 0,
		}
		containerList[i] = cs
	}
	sc := &ScaleContainer{
		name:          msName,
		containerList: containerList,
		data:          msData,
	}
	return sc, nil
}

func (sc *ScaleContainer) StartMicroservice(mNum int) error {
	if sc.data.Scale < mNum {
		return fmt.Errorf(
			"microservice number is over of scale "+
				"(name: %s, scale: %d, request: %d)",
			sc.name, sc.data.Scale, mNum)
	}
	if _, ok := sc.containerList[mNum]; !ok {
		return fmt.Errorf("microservice does not exists (name: %s, number:%d)", sc.name, mNum)
	}
	if sc.containerList[mNum].NumOfUpState > 0 && !sc.data.Multiple {
		return fmt.Errorf(
			"microservice already started, multiple service is not allowed (name: %s, scale: %d, request: %d)",
			sc.name, sc.data.Scale, mNum)
	}
	if err := sc.containerList[mNum].StartProcess(); err != nil {
		return err
	}
	sc.containerList[mNum].NumOfUpState += 1
	return nil
}

func (sc *ScaleContainer) StartAllMicroservice() error {
	var errList []string
	for i := 1; i <= len(sc.containerList); i++ {
		if err := sc.StartMicroservice(i); err != nil {
			errList = append(errList, err.Error())
		}
	}
	if len(errList) != 0 {
		errStr := strings.Join(errList, "\n")
		return fmt.Errorf("StartAllMicroservice is failed :\n%s", errStr)
	}
	return nil
}

func (sc *ScaleContainer) StopMicroservice(mNum int) error {
	if sc.data.Scale < mNum {
		return fmt.Errorf(
			"microservice number is over of scale (name: %s, scale: %d, request: %d)",
			sc.name, sc.data.Scale, mNum)
	}
	if _, ok := sc.containerList[mNum]; !ok {
		return fmt.Errorf("microservice does not exists (name: %s, number:%d)", sc.name, mNum)
	}
	if sc.containerList[mNum].NumOfUpState == 0 {
		return fmt.Errorf(
			"microservice is already finished (name: %s, scale: %d, request: %d)",
			sc.name, sc.data.Scale, mNum)
	}
	if err := sc.containerList[mNum].StopAllProcess(); err != nil {
		return err
	}
	sc.containerList[mNum].NumOfUpState = 0
	return nil
}

func (sc *ScaleContainer) StopAllMicroservice() error {
	var errList []string
	for i := 1; i <= len(sc.containerList); i++ {
		if err := sc.StopMicroservice(i); err != nil {
			errList = append(errList, err.Error())
		}
	}
	if len(errList) != 0 {
		errStr := strings.Join(errList, "\n")
		return fmt.Errorf("StopAllMicroservice is failed :\n%s", errStr)
	}
	return nil
}

func (sc *ScaleContainer) StopMicroserviceForInit(mNum int) error {
	if err := sc.containerList[mNum].StopAllProcess(); err != nil {
		return err
	}
	sc.containerList[mNum].NumOfUpState = 0
	return nil
}

func (sc *ScaleContainer) StopAllMicroserviceForInit() error {
	var errList []string
	for i := 1; i <= len(sc.containerList); i++ {
		if err := sc.StopMicroserviceForInit(i); err != nil {
			errList = append(errList, err.Error())
		}
	}
	if len(errList) != 0 {
		errStr := strings.Join(errList, "\n")
		return fmt.Errorf("StopAllMicroserviceForInit is failed :\n%s", errStr)
	}
	return nil
}
