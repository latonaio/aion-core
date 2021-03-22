// Copyright (c) 2019-2020 Latona. All rights reserved.

package k8s

import (
	"strconv"
	"strings"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	apiV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	hostPathIndex         = 0
	containerPathIndex    = 1
	mountPropagationIndex = 2
)

type Pod struct {
	name                string
	serviceName         string
	tag                 string
	number              int
	command             []string
	ports               []*config.PortConfig
	env                 map[string]string
	volumeMountPathList []string
	serviceAccount      string
	privileged          bool
	k8s                 *k8sResource
	TargetNode          string
}

func NewPod(
	serviceName string, tag string, number int, command []string, ports []*config.PortConfig, env map[string]string, volumeMountPathList []string,
	serviceAccount string, privileged bool, k8s *k8sResource, targetNode string) *Pod {

	return &Pod{
		name:                k8s.getLabelName(serviceName, number),
		serviceName:         serviceName,
		tag:                 tag,
		number:              number,
		command:             command,
		ports:               ports,
		env:                 env,
		volumeMountPathList: volumeMountPathList,
		serviceAccount:      serviceAccount,
		privileged:          privileged,
		k8s:                 k8s,
		TargetNode:          targetNode,
	}
}

func (p *Pod) config() apiV1.PodTemplateSpec {
	return apiV1.PodTemplateSpec{
		ObjectMeta: metaV1.ObjectMeta{
			Labels: p.k8s.getLabelMap(p.serviceName, p.number),
		},
		Spec: apiV1.PodSpec{
			Hostname:              p.name,
			ShareProcessNamespace: boolPrt(true),
			ServiceAccountName:    p.serviceAccount,
			ImagePullSecrets: []apiV1.LocalObjectReference{
				{
					Name: p.k8s.registrySecret,
				},
			},
			Containers: []apiV1.Container{
				p.getContainer(),
				p.getEnvoyContainer(),
			},
			Volumes:      p.getVolumeList(),
			NodeSelector: p.getNodeSelector(),
		},
	}
}

func (p *Pod) getContainer() apiV1.Container {
	return apiV1.Container{
		Name:            p.name,
		Image:           p.k8s.repositoryPrefix + "/" + p.serviceName + ":" + p.tag,
		ImagePullPolicy: apiV1.PullIfNotPresent,
		Command:         p.command,
		SecurityContext: &apiV1.SecurityContext{
			Privileged: &p.privileged,
		},
		Ports:        p.getPortList(),
		Env:          p.getEnvList(),
		VolumeMounts: p.getVolumeMountList(),
	}
}

func (p *Pod) getEnvoyContainer() apiV1.Container {
	return apiV1.Container{
		Name:  "envoy",
		Image: p.k8s.repositoryPrefix + "/envoy:latest",
		Command: []string{
			"/usr/local/bin/envoy",
		},
		Args: []string{
			"--config-path", "/etc/envoy/envoy.yaml",
			"-l", "debug", // for debug
		},
		ImagePullPolicy: apiV1.PullIfNotPresent,
		Resources: apiV1.ResourceRequirements{
			Limits: apiV1.ResourceList{
				apiV1.ResourceMemory: resource.MustParse("512Mi"),
			},
			Requests: apiV1.ResourceList{
				apiV1.ResourceMemory: resource.MustParse("64Mi"),
			},
		},
		Ports: []apiV1.ContainerPort{
			{
				Name:          "envoy-admin",
				ContainerPort: 10001,
			},
		},
		VolumeMounts: []apiV1.VolumeMount{
			{
				Name:      "envoy",
				MountPath: "/etc/envoy",
			},
		},
	}
}

func (p *Pod) getPortList() []apiV1.ContainerPort {
	var portConfigList []apiV1.ContainerPort

	for _, port := range p.ports {
		portConfig := apiV1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.Port,
			Protocol:      apiV1.Protocol(port.Protocol),
		}
		portConfigList = append(portConfigList, portConfig)
	}

	return portConfigList
}

func (p *Pod) getEnvList() []apiV1.EnvVar {
	var envConfList []apiV1.EnvVar

	envConfList = append(envConfList, apiV1.EnvVar{
		Name:  "SERVICE_NAME",
		Value: p.serviceName,
	})

	for key, value := range p.env {
		envConf := apiV1.EnvVar{
			Name:  key,
			Value: value,
		}
		envConfList = append(envConfList, envConf)
	}

	return envConfList
}

func (p *Pod) getVolumeMountList() []apiV1.VolumeMount {
	var volumeMountList []apiV1.VolumeMount

	volumeMountList = append(volumeMountList, apiV1.VolumeMount{
		Name:      "aion-data",
		MountPath: p.k8s.aionDataPath,
	})

	for key, value := range p.volumeMountPathList {
		sValue := strings.Split(value, ":")
		hostPath := sValue[hostPathIndex]
		mountPropagationType := apiV1.MountPropagationNone
		if len(sValue) > mountPropagationIndex && sValue[mountPropagationIndex] == "Bidirectional" {
			mountPropagationType = apiV1.MountPropagationBidirectional
		}

		volumeMount := apiV1.VolumeMount{
			Name:             "data-" + strconv.Itoa(key),
			MountPath:        hostPath,
			MountPropagation: &mountPropagationType,
		}
		volumeMountList = append(volumeMountList, volumeMount)
	}

	return volumeMountList
}

func (p *Pod) getVolumeList() []apiV1.Volume {
	var volumeList []apiV1.Volume

	volumeList = []apiV1.Volume{
		{
			Name: "aion-data",
			VolumeSource: apiV1.VolumeSource{
				HostPath: &apiV1.HostPathVolumeSource{
					Path: p.getHostAionDataPath(),
				},
			},
		},
		{
			Name: "envoy",
			VolumeSource: apiV1.VolumeSource{
				ConfigMap: &apiV1.ConfigMapVolumeSource{
					LocalObjectReference: apiV1.LocalObjectReference{
						Name: "envoy-config-" + p.k8s.getLabelName(p.serviceName, p.number),
					},
				},
			},
		},
	}

	for key, value := range p.volumeMountPathList {
		containerPath := strings.Split(value, ":")[containerPathIndex]

		volume := apiV1.Volume{
			Name: "data-" + strconv.Itoa(key),
			VolumeSource: apiV1.VolumeSource{
				HostPath: &apiV1.HostPathVolumeSource{
					Path: containerPath,
				},
			},
		}
		volumeList = append(volumeList, volume)
	}

	return volumeList
}

func (p *Pod) getHostAionDataPath() string {
	dataPathList := strings.Split(p.k8s.aionDataPath, "/")
	hostDataPath := ""
	for i, path := range dataPathList {
		if len(dataPathList)-1 == i {
			hostDataPath += p.k8s.namespace + "/"
		}
		hostDataPath += path + "/"
	}

	return hostDataPath
}

func (p *Pod) getNodeSelector() map[string]string {
	log.Printf("pod:%v,nodeSector:%v \n", p.name, p.TargetNode)
	if p.TargetNode != "" {
		return map[string]string{"kubernetes.io/hostname": p.TargetNode}
	}
	return nil
}
