// Copyright (c) 2019-2020 Latona. All rights reserved.

package app

import (
	"bitbucket.org/latonaio/aion-core/config"
	"path"
)

var (
	aionHome = "../../../test"
	dataDir  = path.Join(aionHome, "test_data")

	// TODO: docker Config
	// microservicePattern = map[string]bool{"directory": false, "docker": true}
	microservicePattern = map[string]bool{"directory": false}
	testEnv             = map[string]string{"test": "env"}
	microserviceData    = &config.Microservice{
		Command:     "python3 test.py",
		NextService: make(map[string][]*config.NextServiceSetting),
		Scale:       2,
		Env:         testEnv,
		Always:      false,
		Position:    "Runtime",
		Multiple:    false,
		Docker:      false,
		Startup:     false,
		Interval:    0,
	}
	normalMicroserviceDataList = map[string]*config.Microservice{
		"test": microserviceData}
	abnormalMicroserviceDataList = map[string]*config.Microservice{
		"a": microserviceData}
)
