// Copyright (c) 2019-2020 Latona. All rights reserved.

package config

import (
	"fmt"
	"testing"
)

const (
	dataDir        = "./../test/yaml/"
	normalYamlPath = dataDir + "normal_config.yml"
)

func Test_LoadConfig_Normal_001(t *testing.T) {
	c, err := LoadConfigFromDirectory(normalYamlPath, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, val := range c.GetDeviceList() {
		if val.AionHome != DefaultAionHome {
			t.Errorf("cant set default val")
		}
		if val.Username != DefaultUsername {
			t.Errorf("cant set default val")
		}
		if val.Password != DefaultPassword {
			t.Errorf("cant set default val")
		}
		if val.SSHPort != DefaultSSHPort {
			t.Errorf("cant set default val")
		}
	}

	fmt.Printf("%#v\n", c)
	fmt.Printf("%#v\n", c.Aion.Microservices["test"])
	fmt.Printf("%#v\n", c.Aion.Microservices["test"].Command)
	fmt.Printf("%#v\n", c.Aion.Microservices["test"].VolumeMountPathList)
	fmt.Printf("%#v\n", c.Aion.Microservices["test"].NextService["default"].NextServiceSetting)
	/*
		for _, ms := range c.GetMicroserviceList() {
			if len(ms.Env) != 0 {
				if len(ms.Env) != 2 {
					t.Errorf("envirnment length is invalid: %d", len(ms.Env))
				}
				for _, val := range ms.Env {
					if val != "test" {
						t.Errorf("cant get environment value")
					}
				}
			}
			if len(ms.NextService) != 0 {
				if len(ms.NextService) != 1 {
					t.Errorf("next Project length is invalid: %d", len(ms.NextService))
				}
				for _, nService := range ms.NextService {
					if len(nService.NextService) != 2 {
						t.Errorf("cant get next Project list")
					}
					for _, val := range nService {
						if val.NextServiceName == "" {
							t.Errorf("cant set next Project name")
						}
						if val.NumberPattern == "" {
							t.Errorf("cant set next process pattern")
						}
					}
				}
			}
		}
	*/
}
