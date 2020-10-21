// Copyright (c) 2019-2020 Latona. All rights reserved.

package k8s

import (
	"fmt"
	"io/ioutil"
	"strings"
	"unsafe"

	"bitbucket.org/latonaio/aion-core/pkg/log"
	_ "bitbucket.org/latonaio/aion-core/statik"
	"github.com/rakyll/statik/fs"
	apiV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ConfigMap struct {
	serviceName string
	name        string
	number      int
	configMap   v1.ConfigMapInterface
	k8s         *k8sResource
}

func NewConfigMap(serviceName string, number int, k8s *k8sResource) *ConfigMap {
	return &ConfigMap{
		serviceName: serviceName,
		name:        "envoy-config-" + k8s.getLabelName(serviceName, number),
		number:      number,
		configMap:   k8s.client.CoreV1().ConfigMaps(k8s.namespace),
		k8s:         k8s,
	}
}

func (c *ConfigMap) Apply() error {
	config, err := c.config()
	if err != nil {
		return fmt.Errorf("[k8s] create confg is failed: %v", err)
	}

	if _, err := c.configMap.Get(c.k8s.ctx, c.name, metaV1.GetOptions{}); err != nil {
		if _, err := c.configMap.Create(c.k8s.ctx, config, metaV1.CreateOptions{}); err != nil {
			return fmt.Errorf("[k8s] create confg map is failed: %v", err)
		}
		log.Printf("[k8s] Created config map %s", c.name)
	} else {
		if _, err := c.configMap.Update(c.k8s.ctx, config, metaV1.UpdateOptions{}); err != nil {
			return fmt.Errorf("[k8s] update config map is failed: %v", err)
		}
		log.Printf("[k8s] Updated config map %s", c.name)
	}

	return nil
}

func (c *ConfigMap) Delete() error {
	name := "envoy-config-" + c.k8s.getLabelName(c.serviceName, c.number)
	client := c.k8s.client.CoreV1().ConfigMaps(c.k8s.namespace)
	policy := metaV1.DeletePropagationForeground

	if err := client.Delete(
		c.k8s.ctx, name, metaV1.DeleteOptions{PropagationPolicy: &policy}); err != nil {
		return fmt.Errorf("[k8s] Delete config map is failed: %v", err)
	}

	log.Printf("[k8s] Deleted config map %s", name)
	return nil
}

func (c *ConfigMap) config() (*apiV1.ConfigMap, error) {
	statikFs, err := fs.New()
	if err != nil {
		return nil, err
	}
	fp, err := statikFs.Open("/envoy.yaml")
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	conf, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	strConf := strings.ReplaceAll(*(*string)(unsafe.Pointer(&conf)), "{MICROSERVICE_NAME}", c.serviceName)

	return &apiV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name: c.name,
		},
		Data: map[string]string{
			"envoy.yaml": strConf,
		},
	}, nil
}