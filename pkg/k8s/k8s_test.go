// Copyright (c) 2019-2020 Latona. All rights reserved.

package k8s

import (
	"bitbucket.org/latonaio/aion-core/config"
	"context"
	"fmt"
	"github.com/avast/retry-go"
	apiV1 "k8s.io/api/core/v1"
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

var (
	aionHome, _ = filepath.Abs(".")
	ports       = []*config.PortConfig{
		{
			Name:     "web",
			Port:     80,
			Protocol: apiV1.ProtocolTCP,
		},
	}
)

func TestController_DeploymentApply(t *testing.T) {
	ctx := context.TODO()
	k, err := NewK8s(ctx, aionHome)
	if err != nil {
		t.Fatal(err)
	}
	// normal 001 : normal
	t.Run("normal_001", func(t *testing.T) {
		labelName, err := k.DeploymentApply("nginx", 1, ports, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		if labelName == "" {
			t.Fatal("cant get deployment name")
		}
		if err := k.DeploymentDelete(labelName); err != nil {
			t.Error(err)
		}
	})
	// noemal 002 : deployment already exists
	t.Run("normal_002", func(t *testing.T) {
		labelName, err := k.DeploymentApply("nginx", 1, ports, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		defer k.DeploymentDelete(labelName)

		labelName, err = k.DeploymentApply("nginx", 1, ports, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		if labelName == "" {
			t.Fatal("cant get deployment name")
		}
	})
}

func TestController_JobApply(t *testing.T) {
	ctx := context.TODO()
	k, err := NewK8s(ctx, aionHome)
	if err != nil {
		t.Fatal(err)
	}
	ports := []*config.PortConfig{
		{
			Name:     "web",
			Port:     80,
			Protocol: apiV1.ProtocolTCP,
		},
	}
	// normal 001 : normal
	t.Run("normal_001", func(t *testing.T) {
		labelName, err := k.JobApply("nginx", 1, ports, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		if labelName == "" {
			t.Fatal("cant get deployment name")
		}
		if err := k.JobDelete(labelName); err != nil {
			t.Error(err)
		}
	})
	// noemal 002 : deployment already exists
	t.Run("normal_002", func(t *testing.T) {
		labelName, err := k.JobApply("nginx", 1, ports, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		defer k.JobDelete(labelName)

		labelName, err = k.JobApply("nginx", 1, ports, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		if labelName == "" {
			t.Fatal("cant get deployment name")
		}
	})
}

func TestController_ConfigMapApply(t *testing.T) {
	ctx := context.TODO()
	k, err := NewK8s(ctx, aionHome)
	if err != nil {
		t.Fatal(err)
	}
	// normal 001 : normal
	t.Run("normal_001", func(t *testing.T) {
		configMapName, err := k.ConfigMapApply("nginx", 1)
		if err != nil {
			t.Fatal(err)
		}
		if configMapName == "" {
			t.Fatal("cant get deployment name")
		}
		if err := k.ConfigMapDelete(configMapName); err != nil {
			t.Error(err)
		}
	})
	// noemal 002 : deployment already exists
	t.Run("normal_002", func(t *testing.T) {
		configMapName, err := k.ConfigMapApply("nginx", 1)
		if err != nil {
			t.Fatal(err)
		}
		defer k.ConfigMapDelete(configMapName)

		if _, err = k.ConfigMapApply("nginx", 1); err != nil {
			t.Fatal(err)
		}
		if configMapName == "" {
			t.Fatal("cant get deployment name")
		}
	})
}

func TestController_ServiceApply(t *testing.T) {
	ctx := context.TODO()
	k, err := NewK8s(ctx, aionHome)
	if err != nil {
		t.Fatal(err)
	}
	// normal 001 : normal
	t.Run("normal_001", func(t *testing.T) {
		serviceName, err := k.ServiceApply("nginx", 1, ports)
		if err != nil {
			t.Fatal(err)
		}
		if serviceName == "" {
			t.Fatal("cant get deployment name")
		}
		if err := k.ServiceDelete(serviceName); err != nil {
			t.Error(err)
		}
	})
	// normal 002 : deployment already exists
	t.Run("normal_002", func(t *testing.T) {
		serviceName, err := k.ServiceApply("nginx", 1, ports)
		if err != nil {
			t.Fatal(err)
		}
		defer k.ServiceDelete(serviceName)

		if _, err = k.ServiceApply("nginx", 1, ports); err != nil {
			t.Fatal(err)
		}
		if serviceName != "nginx-001" {
			t.Fatalf("cant get expect value (get: %s, expected: %s)", serviceName, "nginx-001")
		}
	})
}

func TestController_IngressSet(t *testing.T) {
	ctx := context.TODO()
	// normal 001 : normal
	t.Run("normal_001", func(t *testing.T) {
		k, err := NewK8s(ctx, aionHome)
		if err != nil {
			t.Fatal(err)
		}
		defer k.Close()
		configName, err := k.ConfigMapApply("nginx", 1)
		if err != nil {
			t.Fatal(err)
		}
		defer k.ConfigMapDelete(configName)

		deploymentName, err := k.DeploymentApply("nginx", 1, ports, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		defer k.DeploymentDelete(deploymentName)
		serviceName, err := k.ServiceApply("nginx", 1, ports)
		if err != nil {
			t.Fatal(err)
		}
		if serviceName == "" {
			t.Fatal("cant get deployment name")
		}
		defer k.ServiceDelete(serviceName)

		if err := k.IngressSet(serviceName, ports); err != nil {
			t.Fatal(err)
		}
		url := "http://localhost" + "/" + serviceName + "/" + ports[0].Name
		req, err := http.NewRequest(
			http.MethodGet, url, nil)
		if err != nil {
			t.Fatalf("cant open port : %d, %v", ports[0].Port, err)
		}
		if err := retry.Do(func() error {
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("cant get response : %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("cant get sussess status (url: %s): %d", url, resp.StatusCode)
			}
			return nil
		},
			retry.Delay(time.Second), retry.Attempts(3)); err != nil {
			t.Error(err)
		}
	})
}
