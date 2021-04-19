// Copyright (c) 2019-2020 Latona. All rights reserved.

package app

import (
	"path"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/proto/servicepb"
)

var (
	aionHome = "../../../test"
	dataDir  = path.Join(aionHome, "test_data")

	// TODO: docker Config
	// microservicePattern = map[string]bool{"directory": false, "docker": true}
	microservicePattern = map[string]bool{"directory": false}
	testEnv             = map[string]string{"test": "env"}
	microserviceData    = &config.Microservice{
		Command:     []string{"python3 test.py"},
		NextService: make(map[string]*servicepb.NextService),
		Scale:       2,
		Env:         testEnv,
		Always:      false,
		Position:    "Runtime",
		Multiple:    false,
		Startup:     false,
		Interval:    0,
	}
	normalMicroserviceDataList = map[string]*config.Microservice{
		"test": microserviceData}
	abnormalMicroserviceDataList = map[string]*config.Microservice{
		"a": microserviceData}
)
