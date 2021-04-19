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
	c, err := LoadConfigFromDirectory(normalYamlPath)
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
}
