// Copyright (c) 2019-2020 Latona. All rights reserved.
package app

import (
	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/devices"
	"bitbucket.org/latonaio/aion-core/internal/microservice"
	"bitbucket.org/latonaio/aion-core/test/mock_config"
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"testing"
)

func getProjectConfigMock(t *testing.T, ctrl *gomock.Controller) *mock_config.MockServiceConfigContainer {
	mock := mock_config.NewMockServiceConfigContainer(ctrl)
	config.GetInstance().ServiceConfigContainer = mock
	return mock
}

func initializeMicroserviceController(aionHome string) *controller {
	return &controller{
		microserviceList: make(map[string]*microservice.ScaleContainer),
		deviceController: &devices.Controller{},
		aionHome:         aionHome,
		watcher:          &Watcher{},
	}
}

func TestNormal001SetMicroservice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	projectConfigMock := getProjectConfigMock(t, ctrl)
	projectConfigMock.EXPECT().GetMicroserviceList().Return(
		normalMicroserviceDataList).AnyTimes()

	for pattern, value := range microservicePattern {
		testName := fmt.Sprintf(
			"TestNormal001SetMicroservice(%s)", pattern)
		t.Run(testName, func(t *testing.T) {
			msc := initializeMicroserviceController(aionHome)
			msList := config.GetInstance().GetMicroserviceList()
			for name, data := range msList {
				data.Docker = value

				if err := msc.setMicroservice(name, data); err != nil {
					t.Fatal(err)
				}
				if len(msc.microserviceList) != 1 {
					t.Errorf("cant set microservice")
				}
			}
		})
	}
}

func TestAbnormal001SetMicroservice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	projectConfigMock := getProjectConfigMock(t, ctrl)
	projectConfigMock.EXPECT().GetMicroserviceList().Return(
		abnormalMicroserviceDataList).AnyTimes()

	for pattern, value := range microservicePattern {
		testName := fmt.Sprintf(
			"TestNormal001SetMicroservice(%s)", pattern)
		t.Run(testName, func(t *testing.T) {
			msc := initializeMicroserviceController(aionHome)
			msList := config.GetInstance().GetMicroserviceList()
			for name, data := range msList {
				data.Docker = value

				if err := msc.setMicroservice(name, data); err == nil {
					t.Error("set invalid microservice, but success to set")
				}
			}
		})
	}
}

func TestNormal001StartMicroservice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	projectConfigMock := getProjectConfigMock(t, ctrl)
	projectConfigMock.EXPECT().GetMicroserviceList().Return(
		normalMicroserviceDataList).AnyTimes()

	msc := initializeMicroserviceController(aionHome)
	msList := config.GetInstance().GetMicroserviceList()
	for name, data := range msList {
		if err := msc.setMicroservice(name, data); err != nil {
			t.Fatal(err)
		}
		if err := msc.startMicroserviceByName(name); err != nil {
			t.Error(err)
		}
	}
}

func TestAbnormal001StartMicroservice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msc := initializeMicroserviceController(aionHome)
	if err := msc.startMicroserviceByName("s"); err == nil {
		t.Errorf("this microservice does not set, but success start microservice")
	}
}

func TestNormal001SetStartMicroservice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	projectConfigMock := getProjectConfigMock(t, ctrl)
	projectConfigMock.EXPECT().GetMicroserviceList().Return(
		normalMicroserviceDataList).AnyTimes()

	msc := initializeMicroserviceController(aionHome)
	if err := msc.setMicroserviceList(); err != nil {
		t.Error(err)
	}

	if len(msc.microserviceList) != 1 {
		t.Errorf("cant set microservice")
	}
}

func TestAbnormal001SetStartMicroservice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	projectConfigMock := getProjectConfigMock(t, ctrl)
	projectConfigMock.EXPECT().GetMicroserviceList().Return(
		abnormalMicroserviceDataList).AnyTimes()

	msc := initializeMicroserviceController(aionHome)
	if err := msc.setMicroserviceList(); err == nil {
		t.Errorf("set invalid microservie, but can set it")
	}
	if len(msc.microserviceList) != 0 {
		t.Errorf(
			"set invalid microservice, but can set it (length: %d)",
			len(msc.microserviceList))
	}
}

func TestNormal001StartMicroserviceController(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	projectConfigMock := getProjectConfigMock(t, ctrl)
	projectConfigMock.EXPECT().GetMicroserviceList().Return(
		normalMicroserviceDataList).AnyTimes()
	projectConfigMock.EXPECT().GetDeviceList().Return(
		map[string]*config.Device{}).AnyTimes()

	env := &Config{
		env: EnvironmentValue{
			AionHome: aionHome,
		},
	}

	if _, err := StartMicroservicesController(ctx, env); err != nil {
		t.Error(err)
	}
}
