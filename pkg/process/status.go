// Copyright (c) 2019-2020 Latona. All rights reserved.
package process

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"fmt"
	"os"
	"sync"
)

type Status struct {
	status      chan *os.ProcessState
	proc        *Process
	stopRequest chan bool
}

type Watcher struct {
	sync.Mutex
	exitProcess       chan *Process
	processStatusList map[int]*Status
}

func NewWatcher() *Watcher {
	watcher := &Watcher{
		exitProcess:       make(chan *Process, 1),
		processStatusList: make(map[int]*Status),
	}
	return watcher
}

func (watcher *Watcher) GetExitProcessCh() chan *Process {
	return watcher.exitProcess
}

func (watcher *Watcher) Add(proc *Process) {
	watcher.Lock()
	defer watcher.Unlock()
	processStatus := &Status{
		status:      make(chan *os.ProcessState, 1),
		proc:        proc,
		stopRequest: make(chan bool, 1),
	}

	// add to watcher process list
	watcher.processStatusList[proc.GetPID()] = processStatus

	// wait to exit process
	go func() {
		log.Printf("start process status watcher (name: %s, PID: %d)",
			proc.GetName(), proc.GetPID())
		state, err := proc.WatchState()
		if err != nil {
			log.Printf(err.Error())
		}
		processStatus.status <- state
	}()

	// wait to exit process or stopping request
	go func() {
		// remove from watcher process list
		defer func() {
			watcher.Lock()
			delete(watcher.processStatusList, proc.GetPID())
			watcher.Unlock()
		}()
		select {
		case status := <-processStatus.status:
			log.Printf("Process exit (name: %s, PID: %d, exitCode: %d)",
				proc.GetName(), proc.GetPID(), status.ExitCode())
			watcher.exitProcess <- proc
			break
		case <-processStatus.stopRequest:
			break
		}
	}()
}

func (watcher *Watcher) Stop(pid int) error {
	// remove from watcher list
	processStatus, ok := watcher.processStatusList[pid]
	if !ok {
		return fmt.Errorf("there is no watcher (PID: %d)", pid)
	}
	log.Printf("stop process status watcher (PID: %d)", pid)
	processStatus.stopRequest <- true
	// wait while stop function
	<-processStatus.status
	return nil
}
