package config

import (
	"io/ioutil"
	"os"

	"bitbucket.org/latonaio/aion-core/proto/devicepb"
	"bitbucket.org/latonaio/aion-core/proto/projectpb"
	"bitbucket.org/latonaio/aion-core/proto/servicepb"
	"gopkg.in/yaml.v2"
)

type YamlAionSetting struct {
	Microservices map[string]*YamlMicroservice `yaml:""`
	Devices       map[string]*YamlDevice       `yaml:",omitempty"`
	DeviceName    string                       `yaml:"deviceName,omitempty"`
	Debug         string                       `yaml:"debug,omitempty"`
}

type YamlDevice struct {
	Addr     string `yaml:""`
	SSHPort  int32  `yaml:"sshPort,omitempty"`
	Username string `yaml:",omitempty"`
	Password string `yaml:",omitempty"`
	AionHome string `yaml:"aionHome,omitempty"`
}

type YamlMicroservice struct {
	Command             []string                      `yaml:",omitempty"`
	NextService         map[string][]*YamlNextService `yaml:"nextService,omitempty"`
	Scale               int32                         `yaml:",omitempty"`
	Env                 map[string]string             `yaml:",omitempty"`
	Position            string                        `yaml:",omitempty"`
	Always              bool                          `yaml:",omitempty"`
	Multiple            bool                          `yaml:",omitempty"`
	Startup             bool                          `yaml:",omitempty"`
	Interval            int32                         `yaml:",omitempty"`
	Ports               []*YamlPortConfig             `yaml:",omitempty"`
	DirPath             string                        `yaml:"directoryPath,omitempty"`
	ServiceAccount      string                        `yaml:"serviceAccount,omitempty"`
	Network             string                        `yaml:",omitempty"`
	Tag                 string                        `yaml:",omitempty"`
	VolumeMountPathList []string                      `yaml:"volumeMountPathList,omitempty"`
	Privileged          bool                          `yaml:",omitempty"`
	WithoutKanban       bool                          `yaml:"withoutKanban,omitempty"`
	TargetNode          string                        `yaml:"targetNode,omitempty"`
	Resources           *YamlResources                `yaml:",omitempty"`
}

type YamlPortConfig struct {
	Name     string `yaml:",omitempty"`
	Protocol string `yaml:",omitempty"`
	Port     int32  `yaml:",omitempty"`
	NodePort int32  `yaml:"nodePort,omitempty"`
}

type YamlNextService struct {
	NextServiceName string `yaml:"name,omitempty"`
	NumberPattern   string `yaml:"pattern,omitempty"`
	NextDevice      string `yaml:"device,omitempty"`
}

type YamlResources struct {
	Requests *YamlResourceConfig `yaml:",omitempty"`
	Limits   *YamlResourceConfig `yaml:",omitempty"`
}

type YamlResourceConfig struct {
	Memory string `yaml:",omitempty"`
	Cpu    string `yaml:",omitempty"`
}

func LoadConfigFromFile(filePath string) (*projectpb.AionSetting, error) {
	aion := &YamlAionSetting{}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(b, aion); err != nil {
		return nil, err
	}
	return aion.mapToGRPCAionSetting(), nil
}

func (as *YamlAionSetting) mapToGRPCAionSetting() *projectpb.AionSetting {
	grpcAionSetting := &projectpb.AionSetting{}
	grpcAionSetting.Devices = map[string]*devicepb.Device{}
	grpcAionSetting.Microservices = map[string]*servicepb.Microservice{}
	if as.Devices != nil {
		for key, value := range as.Devices {
			d := &Device{}
			d.Addr = value.Addr
			d.SSHPort = value.SSHPort
			d.Username = value.Username
			d.Password = value.Password
			d.AionHome = value.AionHome
			grpcAionSetting.Devices[key] = d
		}
	}
	if as.Microservices != nil {
		for key, value := range as.Microservices {
			m := &Microservice{}
			m.NextService = map[string]*servicepb.NextService{}
			m.Command = value.Command
			m.Scale = value.Scale
			m.Env = value.Env
			m.Position = value.Position
			m.Always = value.Always
			m.Multiple = value.Multiple
			m.Startup = value.Startup
			m.Interval = value.Interval
			m.DirPath = value.DirPath
			m.ServiceAccount = value.ServiceAccount
			m.Network = value.Network
			m.Tag = value.Tag
			m.VolumeMountPathList = value.VolumeMountPathList
			m.Privileged = value.Privileged
			m.WithoutKanban = value.WithoutKanban
			m.TargetNode = value.TargetNode
			if value.NextService != nil {
				for k, v := range value.NextService {
					ns := &servicepb.NextService{}
					for _, e := range v {
						nss := &NextServiceSetting{}
						nss.NextServiceName = e.NextServiceName
						nss.NumberPattern = e.NumberPattern
						nss.NextDevice = e.NextDevice
						ns.NextServiceSetting = append(ns.NextServiceSetting, nss)
					}
					m.NextService[k] = ns
				}
			}
			if value.Ports != nil {
				for _, v := range value.Ports {
					p := &PortConfig{}
					p.Name = v.Name
					p.Protocol = v.Protocol
					p.Port = v.Port
					p.NodePort = v.NodePort
					m.Ports = append(m.Ports, p)
				}
			}
			if value.Resources != nil {
				r := &servicepb.Resources{}
				if value.Resources.Requests != nil {
					rr := &ResourceConfig{}
					rr.Memory = value.Resources.Requests.Memory
					rr.Cpu = value.Resources.Requests.Cpu
					r.Requests = rr
				}
				if value.Resources.Limits != nil {
					rl := &ResourceConfig{}
					rl.Memory = value.Resources.Limits.Memory
					rl.Cpu = value.Resources.Limits.Cpu
					r.Limits = rl
				}
				m.Resources = r
			}
			grpcAionSetting.Microservices[key] = m
		}
	}
	grpcAionSetting.DeviceName = as.DeviceName
	grpcAionSetting.Debug = as.Debug
	return grpcAionSetting
}
