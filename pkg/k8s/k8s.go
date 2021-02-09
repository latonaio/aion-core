// Copyright (c) 2019-2020 Latona. All rights reserved.

package k8s

import (
	"context"
	"fmt"
	"strings"

	"bitbucket.org/latonaio/aion-core/pkg/log"
	apiV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var k8sResourceInstance = &k8sResource{}

type K8sResource interface {
	Apply() error
	Delete() error
}

type k8sResource struct {
	ctx              context.Context
	client           *kubernetes.Clientset
	aionDataPath     string
	repositoryPrefix string
	namespace        string
	registrySecret   string
}

type EnvHomeConf struct {
	Home string `envconfig:"HOME"`
}

func New(ctx context.Context, aionDataPath string, repositoryPrefix string, namespace string, registrySecret string) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	k8sResourceInstance = &k8sResource{
		client:           client,
		ctx:              ctx,
		aionDataPath:     aionDataPath,
		repositoryPrefix: repositoryPrefix,
		namespace:        namespace,
		registrySecret:   registrySecret,
	}
	// start pods watcher
	if err := k8sResourceInstance.PodsWatcher(); err != nil {
		return err
	}

	return nil
}

func Get() *k8sResource {
	return k8sResourceInstance
}

func (k *k8sResource) PodsWatcher() error {
	listenerWatcher := cache.NewListWatchFromClient(
		k.client.CoreV1().RESTClient(), string(apiV1.ResourcePods), apiV1.NamespaceAll, fields.Everything())

	var _, watcher = cache.NewInformer(
		listenerWatcher, &apiV1.Pod{}, 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(o interface{}) {
				log.Print("[k8s] Add pod: ", o.(*apiV1.Pod).Name)
			},
			DeleteFunc: func(o interface{}) {
				log.Print("[k8s] Delete pod: ", o.(*apiV1.Pod).Name)
			},
			UpdateFunc: func(old, new interface{}) {
				log.Print("[k8s] Update pod: " + new.(*apiV1.Pod).Name)
			},
		})

	stopCh := make(chan struct{})
	go func() {
		<-k.ctx.Done()
		close(stopCh)
	}()
	go watcher.Run(stopCh)
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
func boolPrt(b bool) *bool    { return &b }

func (k *k8sResource) getLabelName(serviceName string, number int, targetNode string) string {
	t := strings.Split(serviceName, "/")
	t = strings.Split(t[len(t)-1], ":")
	if targetNode == "" {
		return fmt.Sprintf("%s-%03d", t[0], number)
	}
	return fmt.Sprintf("%s-%03d-%s", t[0], number, targetNode)
}

func (k *k8sResource) getLabelNameWithoutTargetNode(serviceName string, number int) string {
	t := strings.Split(serviceName, "/")
	t = strings.Split(t[len(t)-1], ":")
	return fmt.Sprintf("%s-%03d", t[0], number)
}

func (k *k8sResource) getLabelMap(serviceName string, number int) map[string]string {
	return map[string]string{
		"run": k.getLabelNameWithoutTargetNode(serviceName, number),
	}
}

func (k *k8sResource) getObjectMeta(serviceName string, number int, targetNode string) metaV1.ObjectMeta {
	return metaV1.ObjectMeta{
		Labels:    k.getLabelMap(serviceName, number),
		Name:      k.getLabelName(serviceName, number, targetNode),
		Namespace: k.namespace,
	}
}
