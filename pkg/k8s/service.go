// Copyright (c) 2019-2020 Latona. All rights reserved.

package k8s

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"github.com/avast/retry-go"
	apiV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Service struct {
	serviceName string
	name        string
	service     v1.ServiceInterface
	number      int
	ports       []*config.PortConfig
	network     string
	k8sEnv      *K8sEnv
}

func NewService(serviceName string, number int, ports []*config.PortConfig, network string, k8sEnv *K8sEnv) *Service {
	return &Service{
		serviceName: serviceName,
		name:        fmt.Sprintf("%s-srv", getLabelName(serviceName, number)),
		service:     GetClient().CoreV1().Services(k8sEnv.Namespace),
		number:      number,
		ports:       ports,
		network:     network,
		k8sEnv:      k8sEnv,
	}
}

func (s *Service) Apply() error {
	svConfig := s.config()
	ctx := context.Background()

	if _, err := s.service.Get(ctx, s.name, metaV1.GetOptions{}); err != nil {
		result, err := s.service.Create(ctx, svConfig, metaV1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("[k8s] Created service is failed: %v", err)
		}
		log.Printf("[k8s] Created service %s", result.GetObjectMeta().GetName())
	} else {
		if err := s.Delete(); err != nil {
			return err
		}
		result, err := s.service.Create(ctx, svConfig, metaV1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("[k8s] Created service is failed: %v", err)
		}
		log.Printf("[k8s] Deleted & Created service %s", result.GetObjectMeta().GetName())
	}

	return nil
}

func (s *Service) Delete() error {
	name := s.getLabelName(s.serviceName, s.number)
	ctx := context.Background()
	if err := s.service.Delete(ctx, name, metaV1.DeleteOptions{}); err != nil {
		return fmt.Errorf("[k8s] Delete Service is failed: %v", err)
	}

	const connRetryCount = 30
	if err := retry.Do(
		func() error {
			if _, err := s.service.Get(ctx, s.name, metaV1.GetOptions{}); err != nil {
				log.Printf("[k8s] Deleted service %s", name)
				return nil
			}
			return fmt.Errorf("[k8s] Service is not deleted")
		},
		retry.DelayType(func(n uint, config *retry.Config) time.Duration {
			log.Printf("[k8s] Retry to check service is deleted")
			return 2 * time.Second
		}),
		retry.Attempts(connRetryCount),
	); err != nil {
		return err
	}

	return nil
}

func (s *Service) config() *apiV1.Service {
	return &apiV1.Service{
		TypeMeta: metaV1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Labels:    getLabelMap(s.serviceName, s.number),
			Name:      s.name,
			Namespace: s.k8sEnv.Namespace,
		},
		Spec: apiV1.ServiceSpec{
			Type:         s.getServiceType(),
			Ports:        s.getPortConfigList(),
			Selector:     getLabelMap(s.serviceName, s.number),
			TopologyKeys: s.getTopologyKeys(),
		},
		Status: apiV1.ServiceStatus{},
	}
}

func (s *Service) getLabelName(serviceName string, number int) string {
	return fmt.Sprintf("%s-srv", getLabelName(serviceName, number))
}

func (s *Service) getPortConfigList() []apiV1.ServicePort {
	var portConfigList []apiV1.ServicePort

	// open port about envoy admin
	portConfigList = append(portConfigList, apiV1.ServicePort{
		Name:     "envoy-admin",
		Protocol: apiV1.ProtocolTCP,
		Port:     10001,
	})

	// open port by microservice
	for _, port := range s.ports {
		var portConfig apiV1.ServicePort

		switch s.network {
		case "NodePort":
			portConfig = apiV1.ServicePort{
				Name:       port.Name,
				Protocol:   apiV1.Protocol(port.Protocol),
				Port:       port.Port,
				TargetPort: intstr.FromInt(int(port.Port)),
				NodePort:   port.NodePort,
			}
			portConfigList = append(portConfigList, portConfig)
		default:
			portConfig = apiV1.ServicePort{
				Name:       port.Name,
				Protocol:   apiV1.Protocol(port.Protocol),
				Port:       port.Port,
				TargetPort: intstr.FromInt(int(port.Port)),
			}
			portConfigList = append(portConfigList, portConfig)
		}
	}

	return portConfigList
}

func (s *Service) getServiceType() apiV1.ServiceType {
	var serviceType apiV1.ServiceType

	switch s.network {
	case "NodePort":
		serviceType = apiV1.ServiceTypeNodePort
	default:
		serviceType = apiV1.ServiceTypeClusterIP
	}

	return serviceType
}
func (s *Service) getTopologyKeys() []string {
	return []string{"kubernetes.io/hostname"}
}
