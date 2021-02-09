// Copyright (c) 2019-2020 Latona. All rights reserved.
package microservice

import (
	"fmt"
	"strings"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/k8s"
)

type ContainerMicroservice struct {
	K8s       k8s.K8sResource
	ConfigMap *k8s.ConfigMap
	Service   *k8s.Service
}

func NewContainerMicroservice(msName string, data *config.Microservice, msNum int) Container {
	var k k8s.K8sResource
	env := make(map[string]string)
	for k, v := range data.Env {
		env[k] = v
	}

	if !data.Always {
		k = k8s.NewJob(
			msName,
			data.Tag,
			msNum,
			data.Command,
			data.Ports,
			env,
			data.VolumeMountPathList,
			data.ServiceAccount,
			data.Privileged,
			k8s.Get(),
			data.TargetNode,
		)
	} else {
		k = k8s.NewDeployment(
			msName,
			data.Tag,
			msNum,
			data.Command,
			data.Ports,
			env,
			data.VolumeMountPathList,
			data.ServiceAccount,
			data.Privileged,
			k8s.Get(),
			data.TargetNode,
		)
	}

	return &ContainerMicroservice{
		K8s:       k,
		ConfigMap: k8s.NewConfigMap(msName, msNum, k8s.Get(), data.TargetNode),
		Service:   k8s.NewService(msName, msNum, data.Ports, data.Network, k8s.Get()),
	}
}

func (ms *ContainerMicroservice) StartProcess() error {
	if err := ms.ConfigMap.Apply(); err != nil {
		return fmt.Errorf("[start service] failed :%v", err)
	}

	if err := ms.K8s.Apply(); err != nil {
		return fmt.Errorf("[start service] failed :%v", err)
	}

	if err := ms.Service.Apply(); err != nil {
		return fmt.Errorf("[start service] failed :%v", err)
	}

	return nil
}

func (ms *ContainerMicroservice) StopAllProcess() error {
	var errStr []string

	if err := ms.Service.Delete(); err != nil {
		errStr = append(errStr, err.Error())
	}

	if err := ms.K8s.Delete(); err != nil {
		errStr = append(errStr, err.Error())
	}

	if err := ms.ConfigMap.Delete(); err != nil {
		errStr = append(errStr, err.Error())
	}

	if len(errStr) != 0 {
		return fmt.Errorf("[Stop job service] failed: \n%s", strings.Join(errStr, "\n"))
	}

	return nil
}
