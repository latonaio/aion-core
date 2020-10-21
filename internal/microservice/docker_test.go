// Copyright (c) 2019-2020 Latona. All rights reserved.
package microservice

import (
	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/k8s"
	"context"
	"fmt"
	"github.com/avast/retry-go"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"testing"
	"time"
)

var (
	aionHomePath         = "/var/lib/aion"
	testMicroserviceData = &config.Microservice{
		Command:     "",
		NextService: nil,
		Scale:       0,
		Env:         nil,
		Position:    "",
		Always:      true,
		Multiple:    false,
		Docker:      true,
		Startup:     false,
		Interval:    0,
		Ports: []*config.PortConfig{
			{
				Name:     "web",
				Protocol: v1.ProtocolTCP,
				Port:     80,
			},
		},
	}
)

func TestNormal001StartProcess(t *testing.T) {
	msName := "nginx"
	msNumber := 1
	ctx := context.TODO()
	if err := k8s.InitializeK8s(ctx, aionHomePath); err != nil {
		t.Fatal(err)
	}
	testMicroserviceData.Always = true
	ms := NewDockerMicroservice(msName, testMicroserviceData, msNumber)
	if err := ms.StartProcess(); err != nil {
		t.Error(err)
	}
	url := "http://localhost/nginx-001-srv/web"
	req, err := http.NewRequest(
		http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("cant open port (url: %s) %v", url, err)
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
	}, retry.Delay(time.Second), retry.Attempts(10)); err != nil {
		t.Error(err)
	}
	if err := ms.StopAllProcess(); err != nil {
		t.Error(err)
	}
}
