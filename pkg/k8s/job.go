// Copyright (c) 2019-2020 Latona. All rights reserved.

package k8s

import (
	"fmt"
	"time"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"github.com/avast/retry-go"
	batchV1 "k8s.io/api/batch/v1"
	apiV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/batch/v1"
)

type Job struct {
	name        string
	serviceName string
	pod         *Pod
	job         v1.JobInterface
	k8s         *k8sResource
}

func NewJob(
	serviceName string, tag string, number int, command []string, ports []*config.PortConfig, env map[string]string, volumeMountPathList []string,
	serviceAccount string, privileged bool, k8s *k8sResource) *Job {

	return &Job{
		name:        k8s.getLabelName(serviceName, number),
		serviceName: serviceName,
		job:         k8s.client.BatchV1().Jobs(k8s.namespace),
		k8s:         k8s,
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
			k8s,
		),
	}
}

func (j *Job) Apply() error {
	config := j.config()

	if _, err := j.job.Get(j.k8s.ctx, j.name, metaV1.GetOptions{}); err != nil {
		result, err := j.job.Create(j.k8s.ctx, config, metaV1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("[k8s] apply job is failed: %v", err)
		}
		log.Printf("[k8s] Created Job %s", result.GetObjectMeta().GetName())

	} else {
		if err := j.Delete(); err != nil {
			return err
		}
		const connRetryCount = 10
		if err := retry.Do(
			func() error {
				if _, err := j.job.Get(j.k8s.ctx, j.name, metaV1.GetOptions{}); err != nil {
					log.Printf("[k8s] Duplicate job is deleted")
					return nil
				}
				return fmt.Errorf("[k8s] Duplicate job is not deleted")
			},
			retry.DelayType(func(n uint, config *retry.Config) time.Duration {
				log.Printf("[k8s] Retry to check duplicated job is deleted")
				return time.Second
			}),
			retry.Attempts(connRetryCount),
		); err != nil {
			return err
		}

		result, err := j.job.Create(j.k8s.ctx, config, metaV1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("[k8s] apply job is failed: %v", err)
		}
		log.Printf("[k8s] Deleted & Created Job %s", result.GetObjectMeta().GetName())
	}

	return nil
}

func (j *Job) Delete() error {
	deletePolicy := metaV1.DeletePropagationForeground
	if err := j.job.Delete(
		j.k8s.ctx, j.name, metaV1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
		return fmt.Errorf("[k8s] Delete job is failed: %v", err)
	}

	log.Printf("[k8s] Deleted job %s", j.name)
	return nil
}

func (j *Job) config() *batchV1.Job {
	podConfig := j.pod.config()
	podConfig.Spec.RestartPolicy = apiV1.RestartPolicyOnFailure

	return &batchV1.Job{
		TypeMeta: metaV1.TypeMeta{
			Kind:       "job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: j.k8s.getObjectMeta(j.serviceName, j.pod.number),
		Spec: batchV1.JobSpec{
			Completions:             int32Ptr(1),
			Parallelism:             int32Ptr(1),
			BackoffLimit:            int32Ptr(1),
			Template:                podConfig,
			TTLSecondsAfterFinished: int32Ptr(2),
		},
	}
}
