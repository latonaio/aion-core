// Copyright (c) 2019-2020 Latona. All rights reserved.

package k8s

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"github.com/avast/retry-go"
	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

type Deployment struct {
	name        string
	serviceName string
	deployment  v1.DeploymentInterface
	pod         *Pod
	k8sEnv      *K8sEnv
}

func NewDeployment(
	serviceName string, tag string, number int, command []string, ports []*config.PortConfig, env map[string]string, volumeMountPathList []string,
	serviceAccount string, privileged bool, k8sEnv *K8sEnv, targetNode string, resources *config.Resources) *Deployment {

	return &Deployment{
		name:        getLabelName(serviceName, number),
		serviceName: serviceName,
		deployment:  GetClient().AppsV1().Deployments(k8sEnv.Namespace),
		k8sEnv:      k8sEnv,
		pod: NewPod(
			serviceName,
			tag,
			number,
			command,
			ports,
			env,
			volumeMountPathList,
			serviceAccount,
			privileged,
			k8sEnv,
			targetNode,
			resources,
		),
	}
}

func (d *Deployment) Apply() error {
	dplConfig := d.config()
	ctx := context.Background()

	if _, err := d.deployment.Get(ctx, d.name, metaV1.GetOptions{}); err != nil {
		result, err := d.deployment.Create(ctx, dplConfig, metaV1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("[k8s] apply deployment is failed: %v", err)
		}
		log.Printf("[k8s] Created deployment %s", result.GetObjectMeta().GetName())
	} else {
		result, err := d.deployment.Update(ctx, dplConfig, metaV1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("[k8s] apply deployment is failed: %v", err)
		}
		log.Printf("[k8s] Updated deployment %s", result.GetObjectMeta().GetName())
	}

	return nil
}

func (d *Deployment) Delete() error {
	deletePolicy := metaV1.DeletePropagationForeground
	ctx := context.Background()
	if err := d.deployment.Delete(
		ctx, d.name, metaV1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
		return fmt.Errorf("[k8s] Delete deployment is failed: %v", err)
	}

	const connRetryCount = 30
	if err := retry.Do(
		func() error {
			if _, err := d.deployment.Get(ctx, d.name, metaV1.GetOptions{}); err != nil {
				log.Printf("[k8s] Deleted deployment %s", d.name)
				return nil
			}
			return fmt.Errorf("[k8s] Deployment is not deleted")
		},
		retry.DelayType(func(n uint, config *retry.Config) time.Duration {
			log.Printf("[k8s] Retry to check deployment is deleted")
			return 2 * time.Second
		}),
		retry.Attempts(connRetryCount),
	); err != nil {
		return err
	}

	return nil
}

func (d *Deployment) config() *appsV1.Deployment {
	return &appsV1.Deployment{
		ObjectMeta: getObjectMeta(d.k8sEnv.Namespace, d.serviceName, d.pod.number, d.pod.TargetNode),
		Spec: appsV1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metaV1.LabelSelector{
				MatchLabels: getLabelMap(d.serviceName, d.pod.number),
			},
			Strategy: appsV1.DeploymentStrategy{
				RollingUpdate: &appsV1.RollingUpdateDeployment{},
			},
			Template: d.pod.config(),
		},
	}
}
