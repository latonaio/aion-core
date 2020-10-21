package kanban

import (
	"bitbucket.org/latonaio/aion-core/pkg/common"
	"bitbucket.org/latonaio/aion-core/pkg/fswatcher"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/golang/protobuf/jsonpb"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func NewFileAdapter(aionDataPath string) Adaptor {
	return &FileAdapter{aionDataPath: aionDataPath}
}

type FileAdapter struct {
	aionDataPath string
	watcher      *fswatcher.FileWatcher
}

func (fa *FileAdapter) getPath(msName string, msNumber int, createDir bool) (string, error) {
	msDataPath := common.GetMsDataPath(fa.aionDataPath, msName, msNumber)
	// create directory
	if _, err := os.Stat(msDataPath); os.IsNotExist(err) {
		if !createDir {
			return "", fmt.Errorf("there is no directory: %s", msDataPath)
		}
		if err := os.Mkdir(msDataPath, 0755); err != nil {
			return "", fmt.Errorf("cant create directory: %s", msDataPath)
		}
	}
	return msDataPath, nil
}

func convertStatusType(statusType StatusType) string {
	switch statusType {
	case StatusType_Before:
		return "B"
	case StatusType_After:
		return "A"
	default:
		return ""
	}
}

func (fa *FileAdapter) WatchKanban(msName string, msNumber int, statusType StatusType) (chan *kanbanpb.StatusKanban, error) {
	dataPath, err := fa.getPath(msName, msNumber, true)
	if err != nil {
		return nil, fmt.Errorf("cant watch kanban, because directory does not exist: %s", dataPath)
	}

	// set watcher
	watcher, err := fswatcher.NewCreateWatcher()
	if err != nil {
		return nil, err
	}

	// start file Watcher
	watcher.AddDirectory(dataPath)
	fa.watcher = watcher

	resCh := make(chan *kanbanpb.StatusKanban)
	go func() {
		for kanbanPath := range watcher.GetFilePathCh() {
			if strings.Split(path.Base(kanbanPath), "_")[0] != convertStatusType(statusType) {
				continue
			}
			kanban, err := fa.ReadKanban(msName, msNumber, statusType)
			if err != nil {
				log.Printf("cant get kanban: %v", err)
				continue
			}
			resCh <- kanban
		}
	}()

	log.Printf("Start Watch Microservice: %s", dataPath)
	return resCh, nil
}

func (fa *FileAdapter) ReadKanban(msName string, msNumber int, statusType StatusType) (*kanbanpb.StatusKanban, error) {
	dataPath, err := fa.getPath(msName, msNumber, false)
	if err != nil {
		return nil, fmt.Errorf("cant read kanban, because directory does not exist: %s", dataPath)
	}

	fileList, err := filepath.Glob(path.Join(dataPath, convertStatusType(statusType)+"*.json"))
	if err != nil {
		return nil, fmt.Errorf("getLatestStatusKanban: %v", err)
	}
	if len(fileList) == 0 {
		return nil, fmt.Errorf("there is no kanban (path: %s)", dataPath)
	}
	sort.Strings(fileList)
	filePath := fileList[0]

	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("there is no file: %v", err)
	}
	kanban := &kanbanpb.StatusKanban{}
	// retry to read file
	if err := retry.Do(
		func() error {
			r, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("cant open file (path: %s): %v", filePath, err)
			}
			defer r.Close()
			if err := jsonpb.Unmarshal(r, kanban); err != nil {
				return fmt.Errorf("cant unmarshal json file (path: %s): %v", filePath, err)
			}
			return nil
		},
		retry.Attempts(3),
	); err != nil {
		return nil, fmt.Errorf("cant read json file: %v", err)
	}

	return kanban, nil
}

func (fa *FileAdapter) WriteKanban(
	msName string, msNumber int, kanban *kanbanpb.StatusKanban, statusType StatusType) error {
	dataPath, err := fa.getPath(msName, msNumber, true)
	if err != nil {
		return fmt.Errorf("cant write kanban, because directory does not exist: %s", dataPath)
	}
	nowDate := common.GetFileNameDatetime()

	count := 0
	filePath := ""
	for {
		count += 1
		fileName := strings.Join(
			[]string{convertStatusType(statusType), nowDate, strconv.Itoa(count), msName + ".json"}, "_")
		filePath = path.Join(dataPath, fileName)
		if _, err := os.Stat(filePath); err != nil {
			break
		}
	}

	// open write file by io
	w, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer w.Close()

	// write kanban
	m := jsonpb.Marshaler{EmitDefaults: true, Indent: "    ", OrigName: true}
	if err := m.Marshal(w, kanban); err != nil {
		return err
	}
	log.Printf("output kanban to file (path: %s)", filePath)

	return nil
}
