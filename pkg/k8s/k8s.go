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

var k8sClient *kubernetes.Clientset = func() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("cannot get config: %v\n", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("cannot get client: %v\n", err)
	}
	return client
}()

type K8sResource interface {
	Apply() error
	Delete() error
}

type K8sEnv struct {
	AionDataPath     string
	RepositoryPrefix string
	Namespace        string
	RegistrySecret   string
}

func newClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetClient() *kubernetes.Clientset {
	return k8sClient
}

func NewK8sEnv(aionDataPath string, repositoryPrefix string, namespace string, registrySecret string) *K8sEnv {
	return &K8sEnv{
		AionDataPath:     aionDataPath,
		RepositoryPrefix: repositoryPrefix,
		Namespace:        namespace,
		RegistrySecret:   registrySecret,
	}
}

func PodsWatcher(ctx context.Context) error {
	client := GetClient()

	listenerWatcher := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(), string(apiV1.ResourcePods), apiV1.NamespaceAll, fields.Everything())

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
		<-ctx.Done()
		close(stopCh)
	}()
	go watcher.Run(stopCh)
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
func boolPrt(b bool) *bool    { return &b }

func getLabelName(serviceName string, number int) string {
	t := strings.Split(serviceName, "/")
	t = strings.Split(t[len(t)-1], ":")
	return fmt.Sprintf("%s-%03d", t[0], number)
}

func getLabelNameWithoutTargetNode(serviceName string, number int) string {
	t := strings.Split(serviceName, "/")
	t = strings.Split(t[len(t)-1], ":")
	return fmt.Sprintf("%s-%03d", t[0], number)
}

func getLabelMap(serviceName string, number int) map[string]string {
	return map[string]string{
		"run": getLabelNameWithoutTargetNode(serviceName, number),
	}
}

func getObjectMeta(namespace string, serviceName string, number int, targetNode string) metaV1.ObjectMeta {
	return metaV1.ObjectMeta{
		Labels:    getLabelMap(serviceName, number),
		Name:      getLabelName(serviceName, number),
		Namespace: namespace,
	}
}
