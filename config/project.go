// Copyright (c) 2019-2020 Latona. All rights reserved.
package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
	apiV1 "k8s.io/api/core/v1"
)

var instance = &Config{
	&AionSetting{},
}

func GetInstance() *Config {
	return instance
}

type Config struct {
	ServiceConfigContainer
}

type ServiceConfigContainer interface {
	GetMicroserviceList() map[string]*Microservice
	GetMicroserviceByName(name string) (*Microservice, error)
	GetNextServiceList(name string, connectionKey string) ([]*NextServiceSetting, error)
	GetDeviceName() string
	GetDeviceList() map[string]*Device
	LoadConfig(confPath string, isDocker bool) error
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
	Microservices map[string]*Microservice `yaml:""`
	Devices       map[string]*Device       `yaml:",omitempty"`
	DeviceName    string                   `yaml:"deviceName,omitempty"`
}

type Microservice struct {
	Command             []string                         `yaml:",omitempty"`
	NextService         map[string][]*NextServiceSetting `yaml:"nextService,omitempty"`
	Scale               int                              `yaml:",omitempty"`
	Env                 map[string]string                `yaml:",omitempty"`
	Position            Position                         `yaml:",omitempty"`
	Always              bool                             `yaml:",omitempty"`
	Multiple            bool                             `yaml:",omitempty"`
	Docker              bool                             `yaml:",omitempty"`
	Startup             bool                             `yaml:",omitempty"`
	Interval            int                              `yaml:",omitempty"`
	Ports               []*PortConfig                    `yaml:",omitempty"`
	DirPath             string                           `yaml:"directoryPath,omitempty"`
	ServiceAccount      string                           `yaml:"serviceAccount,omitempty"`
	Network             string                           `yaml:",omitempty"`
	Tag                 string                           `yaml:",omitempty"`
	VolumeMountPathList []string                         `yaml:"volumeMountPathList,omitempty"`
	Privileged          bool                             `yaml:",omitempty"`
	WithoutKanban       bool                             `yaml:"withoutKanban,omitempty"`
	TargetNode          string                           `yaml:"targetNode,omitempty"`
}

type PortConfig struct {
	Name     string         `yaml:""`
	Protocol apiV1.Protocol `yaml:",omitempty"`
	Port     int32          `yaml:""`
	NodePort int32          `yaml:"nodePort,omitempty"`
}

func (m *Microservice) GetPosition() string {
	return string(m.Position)
}

type NextServiceSetting struct {
	NextServiceName string          `yaml:"name"`
	NumberPattern   msNumberPattern `yaml:"pattern"`
	NextDevice      string          `yaml:"device,omitempty"`
}

type Position string

const (
	Runtime        Position = "Runtime"
	BackendService Position = "BackendService"
	UI             Position = "UI"
)

type msNumberPattern string

const (
	DefaultNumber msNumberPattern = "0"
	Spread        msNumberPattern = "n"
)

// Device ... edge device params
type Device struct {
	Addr     string `yaml:""`
	SSHPort  int16  `yaml:"sshPort,omitempty"`
	Username string `yaml:",omitempty"`
	Password string `yaml:",omitempty"`
	AionHome string `yaml:"aionHome,omitempty"`
}

func GetNextNumber(
	previousNumber int32, msNumberPattern msNumberPattern) int {
	switch msNumberPattern {
	case DefaultNumber:
		return 1
	case Spread:
		return int(previousNumber)
	default:
		return 1
	}
}

func (ya *AionSetting) LoadConfig(configPath string, isDocker bool) error {
	f, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(b, ya); err != nil {
		return err
	}
	if err := ya.setInitializeValue(isDocker); err != nil {
		return err
	}
	return nil
}

func (ya *AionSetting) GetMicroserviceByName(name string) (*Microservice, error) {
	ms, ok := ya.Microservices[name]
	if !ok {
		return nil, fmt.Errorf("there is no microservice: %s", name)
	}
	return ms, nil
}

func (ya *AionSetting) GetNextServiceList(name string, connectionKey string) ([]*NextServiceSetting, error) {
	ms, ok := ya.Microservices[name]
	if !ok {
		return nil, fmt.Errorf("there is no microservice: %s", name)
	}
	if cl, ok := ms.NextService[connectionKey]; ok {
		return cl, nil
		// get default key
	} else if cl, ok := ms.NextService[DefaultConnectionKey]; ok {
		return cl, nil
	}
	return nil, fmt.Errorf(
		"invalid connection key (connectionkey: %s)", connectionKey)
}

func (ya *AionSetting) GetMicroserviceList() map[string]*Microservice {
	return ya.Microservices
}

func (ya *AionSetting) GetDeviceName() string {
	return ya.DeviceName
}

func (ya *AionSetting) GetDeviceList() map[string]*Device {
	return ya.Devices
}

func (ya *AionSetting) setInitializeValue(isDocker bool) error {
	for _, val := range ya.Devices {
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

	for _, msData := range ya.Microservices {
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
			msData.NextService = map[string][]*NextServiceSetting{}
		}
		if msData.Tag == "" {
			msData.Tag = DefaultTag
		}
		if isDocker {
			msData.Docker = !msData.Docker
		}

		for _, nextKey := range msData.NextService {
			for _, nextMs := range nextKey {
				if nextMs.NumberPattern == "" {
					nextMs.NumberPattern = DefaultProcessPattern
				}
			}
		}
	}
	return nil
}
