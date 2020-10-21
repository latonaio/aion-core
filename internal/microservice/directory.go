// Copyright (c) 2019-2020 Latona. All rights reserved.
package microservice

import (
	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/pkg/process"
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"
)

type DirectoryMicroservice struct {
	sync.Mutex
	name           string
	process        map[int]*process.Process
	processWatcher *process.Watcher
	data           *config.Microservice
	number         int
	path           string
}

func NewDirectoryMicroservice(aionHome string, msName string, data *config.Microservice, mNum int) (Container, error) {
	// existence check
	var dirName string
	if data.DirPath != "" {
		dirName = data.DirPath
	} else {
		dirName = msName
	}
	scriptPath := path.Join(aionHome, string(data.Position), dirName)
	log.Printf(scriptPath)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, err
	}

	watcher := process.NewWatcher()
	ms := &DirectoryMicroservice{
		name:           msName,
		process:        make(map[int]*process.Process),
		processWatcher: watcher,
		data:           data,
		number:         mNum,
		path:           scriptPath,
	}
	go ms.watchRestart()

	return ms, nil
}

func (ms *DirectoryMicroservice) watchRestart() {
	for proc := range ms.processWatcher.GetExitProcessCh() {
		ms.processWatcher.Stop(proc.GetPID())
		delete(ms.process, proc.GetPID())

		if ms.data.Always {
			log.Printf("restart process (name: %s)", ms.name)

			if err := ms.StartProcess(); err != nil {
				log.Printf(err.Error())
			}
		}
	}
}

func (ms *DirectoryMicroservice) StartProcess() error {
	// start process
	ms.Lock()
	defer ms.Unlock()
	if !ms.data.Multiple && len(ms.process) == 1 {
		return fmt.Errorf("(ms: %s) multiple process is not allowd", ms.name)
	}

	var envList []string
	for key, val := range ms.data.Env {
		envStr := key + "=" + val
		envList = append(envList, envStr)
	}
	envList = append(envList, "MS_NUMBER="+strconv.Itoa(ms.number))

	proc, err := process.NewProcess(ms.name, ms.data.Command, ms.path, envList)
	if err != nil {
		return err
	}

	ms.process[proc.GetPID()] = proc
	// start process watcher
	ms.processWatcher.Add(proc)

	return nil
}

func (ms *DirectoryMicroservice) StopAllProcess() error {
	ms.Lock()
	log.Printf("(ms: %s) Stop all process", ms.name)
	for pid, proc := range ms.process {
		proc.Stop()
		ms.processWatcher.Stop(pid)
		delete(ms.process, pid)
	}
	ms.Unlock()
	return nil
}
