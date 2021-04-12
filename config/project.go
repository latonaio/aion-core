// Copyright (c) 2019-2020 Latona. All rights reserved.
package config

import (
	"fmt"

	"bitbucket.org/latonaio/aion-core/proto/devicepb"
	"bitbucket.org/latonaio/aion-core/proto/projectpb"
	"bitbucket.org/latonaio/aion-core/proto/servicepb"
)

type ServiceConfigContainer interface {
	GetMicroserviceList() map[string]*Microservice
	GetMicroserviceByName(name string) (*Microservice, error)
	GetNextServiceList(name string, connectionKey string) ([]*NextServiceSetting, error)
	GetDeviceName() string
	GetDebug() string
	GetDeviceList() map[string]*Device
}

const (
	DefaultAionHome       = "/var/lib/aion"
	DefaultUsername       = "aion"
	DefaultPassword       = "aion"
	DefaultSSHPort        = 22
	DefaultProcessPattern = Spread
	DefaultConnectionKey  = "default"
	DefaultPosition       = Runtime
	DefaultScale          = 1
	DefaultServiceAccount = "default"
	DefaultNetwork        = "ClusterIP"
	DefaultTag            = "latest"
)

type AionSetting struct {
	Aion *projectpb.AionSetting
}

type Microservice = servicepb.Microservice

type PortConfig = servicepb.PortConfig

type NextServiceSetting = servicepb.NextServiceSetting

type Resources = servicepb.Resources

type ResourceConfig = servicepb.ResourceConfig

type Device = devicepb.Device

const (
	Runtime        string = "Runtime"
	BackendService string = "BackendService"
	UI             string = "UI"
)

const (
	DefaultNumber string = "0"
	Spread        string = "n"
)

func GetNextNumber(
	previousNumber int32, msNumberPattern string) int {
	switch msNumberPattern {
	case DefaultNumber:
		return 1
	case Spread:
		return int(previousNumber)
	default:
		return 1
	}
}

func LoadConfigFromDirectory(configPath string, isDocker bool) (*AionSetting, error) {
	/*
	 * 本来ならProtocol Bufferで定義されたモデルをそのまま使うべきだが、
	 * 既存のproject.yamlがProtocol Buffetで表現できない
	 * Protocol Bufferでyaml mappingができない
	 * という2つの理由から、yaml読み込み用のモデルを定義して読み込んだ後、
	 * Protocol Bufferで定義されたモデルにマッピングしている。
	 */
	aion, err := LoadConfigFromFile(configPath)
	if err != nil {
		return nil, err
	}
	aionSetting := &AionSetting{aion}
	if err := aionSetting.setInitializeValue(isDocker); err != nil {
		return nil, err
	}
	return aionSetting, nil
}

func LoadConfigFromGRPC(project *projectpb.AionSetting, isDocker bool) (*AionSetting, error) {
	aionSetting := &AionSetting{Aion: project}
	if err := aionSetting.setInitializeValue(isDocker); err != nil {
		return nil, err
	}
	return aionSetting, nil
}

func (ya *AionSetting) GetMicroserviceByName(name string) (*Microservice, error) {
	ms, ok := ya.Aion.Microservices[name]
	if !ok {
		return nil, fmt.Errorf("there is no microservice: %s", name)
	}
	return ms, nil
}

func (ya *AionSetting) GetNextServiceList(name string, connectionKey string) ([]*NextServiceSetting, error) {
	ms, ok := ya.Aion.Microservices[name]
	if !ok {
		return nil, fmt.Errorf("there is no microservice: %s", name)
	}
	if cl, ok := ms.NextService[connectionKey]; ok {
		return cl.NextServiceSetting, nil
		// get default key
	} else if cl, ok := ms.NextService[DefaultConnectionKey]; ok {
		return cl.NextServiceSetting, nil
	}
	return nil, fmt.Errorf(
		"invalid connection key (connectionkey: %s)", connectionKey)
}

func (ya *AionSetting) GetMicroserviceList() map[string]*Microservice {
	return ya.Aion.Microservices
}

func (ya *AionSetting) GetDeviceName() string {
	return ya.Aion.DeviceName
}

func (ya *AionSetting) GetDebug() string {
	return ya.Aion.Debug
}

func (ya *AionSetting) GetDeviceList() map[string]*Device {
	return ya.Aion.Devices
}

func (ya *AionSetting) setInitializeValue(isDocker bool) error {
	for _, val := range ya.Aion.Devices {
		if val.AionHome == "" {
			val.AionHome = DefaultAionHome
		}
		if val.Username == "" {
			val.Username = DefaultUsername
		}
		if val.Password == "" {
			val.Password = DefaultPassword
		}
		if val.SSHPort == 0 {
			val.SSHPort = DefaultSSHPort
		}
	}

	for _, msData := range ya.Aion.Microservices {
		if msData == nil {
			return fmt.Errorf("yaml format is wrong")
		}
		if msData.Position == "" {
			msData.Position = DefaultPosition
		}
		if msData.Scale <= 0 {
			msData.Scale = DefaultScale
		}
		if msData.Network == "" {
			msData.Network = DefaultNetwork
		}
		if msData.ServiceAccount == "" {
			msData.ServiceAccount = DefaultServiceAccount
		}
		if msData.Env == nil {
			msData.Env = map[string]string{}
		}
		if msData.NextService == nil {
			msData.NextService = map[string]*servicepb.NextService{}
		}
		if msData.Tag == "" {
			msData.Tag = DefaultTag
		}
		msData.Docker = isDocker

		for _, nextKey := range msData.NextService {
			for _, nextMs := range nextKey.NextServiceSetting {
				if nextMs.NumberPattern == "" {
					nextMs.NumberPattern = DefaultProcessPattern
				}
			}
		}
	}
	return nil
}
