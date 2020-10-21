package kanban

import (
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

var (
	dataPath            = "./../test/test_data"
	priorServiceName    = "test1"
	priorServiceNumber  = 1
	abnormalServiceName = "abnormal_test"
	nextServiceName     = "test2"
	cJsonSampleName     = "csample.json"
)

func TestFileAdapter_ReadKanban(t *testing.T) {
	fa := FileAdapter{aionDataPath: dataPath}
	// normal case
	kanban, err := fa.ReadKanban(priorServiceName, priorServiceNumber, StatusType_After)
	if err != nil {
		t.Fatal(err)
	}
	if kanban.Services[0] != "test" || !kanban.PriorSuccess {
		t.Errorf("cant get expect value: (get: %s, expect:%s)", kanban.Services[0], "test")
	}

	// abnormal case: there is no file
	kanban, err = fa.ReadKanban(nextServiceName, priorServiceNumber, StatusType_Before)
	if err == nil {
		t.Errorf("there is no kanban, but can read file")
	}
}

func TestFileAdapter_WriteKanban(t *testing.T) {
	fa := FileAdapter{aionDataPath: dataPath}

	kanban := &kanbanpb.StatusKanban{
		StartAt:       "",
		FinishAt:      "",
		Services:      []string{"test"},
		ConnectionKey: "key",
		ProcessNumber: 1,
		PriorSuccess:  true,
		DataPath:      "",
		Metadata:      &_struct.Struct{},
	}

	// normal case : write to next service dir
	if err := fa.WriteKanban(nextServiceName, priorServiceNumber, kanban, StatusType_After); err != nil {
		t.Fatal(err)
	}
	// check exist file
	outputPath := path.Join(dataPath, nextServiceName+"_"+strconv.Itoa(priorServiceNumber))
	defer os.RemoveAll(outputPath)
	fileList, err := filepath.Glob(path.Join(outputPath, "A_*.json"))
	if err != nil {
		t.Error(err)
	}
	if len(fileList) == 0 {
		t.Error("cant create json file")
	}
}

func TestFileAdapter_WatchKanban(t *testing.T) {
	fa := FileAdapter{aionDataPath: dataPath}

	kanban := &kanbanpb.StatusKanban{
		StartAt:       "",
		FinishAt:      "",
		Services:      []string{"test"},
		ConnectionKey: "key",
		ProcessNumber: 1,
		PriorSuccess:  true,
		DataPath:      "",
		Metadata:      &_struct.Struct{},
	}

	ch, err := fa.WatchKanban(nextServiceName, priorServiceNumber, StatusType_After)
	if err != nil {
		t.Fatal(err)
	}
	outputPath := path.Join(dataPath, nextServiceName+"_"+strconv.Itoa(priorServiceNumber))
	defer os.RemoveAll(outputPath)

	// normal case : write to next service dir
	if err := fa.WriteKanban(nextServiceName, priorServiceNumber, kanban, StatusType_After); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ch:
		break
	case <-time.After(time.Millisecond * 500):
		t.Errorf("cant catch kanban by watcher")
	}
}
