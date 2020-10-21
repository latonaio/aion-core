// Copyright (c) 2019-2020 Latona. All rights reserved.
package process

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Process struct {
	Name string
	Cmd  string
	Args []string
	Proc *os.Process
	Env  []string
	PID  int
	Path string
}

func NewProcess(name string, command []string, dirPath string, envList []string) (*Process, error) {
	proc := &Process{
		Name: name,
		Cmd:  command[0],
		Args: command[1:],
		PID:  -1,
		Env:  envList,
		Path: dirPath,
	}
	if err := proc.Start(); err != nil {
		return nil, err
	}

	return proc, nil
}

func (proc *Process) Start() error {
	if err := proc.IsAlive(); err == nil {
		return fmt.Errorf("%s is already started", proc.Name)
	}
	// #nosec G204
	cmd := exec.Command(proc.Cmd, proc.Args...)
	cmd.Dir = proc.Path

	// TODO: output to logger service
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if len(proc.Env) != 0 {
		cmd.Env = append(os.Environ(), proc.Env...)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	proc.Proc = cmd.Process
	proc.PID = cmd.Process.Pid
	log.Printf("process is started from service-broker (name: %s, PID: %d)",
		proc.Name, proc.PID)

	return nil
}

func (proc *Process) GetPID() int {
	return proc.PID
}

func (proc *Process) GetName() string {
	return proc.Name
}

func (proc *Process) stop(isKill bool) error {
	if err := proc.IsAlive(); err != nil {
		return err
	}

	killed := false
	// send sigterm
	if !isKill {
		if err := proc.Proc.Signal(syscall.SIGTERM); err == nil {
			killed = true
		}
	}
	// send kill when request to force stop or cant kill by sigterm
	if isKill || !killed {
		if err := proc.Proc.Signal(syscall.SIGKILL); err != nil {
			return err
		}
	}
	// release process
	if err := proc.Proc.Release(); err != nil {
		return err
	}
	log.Printf("process is terminated from service-broker (name: %s, PID: %d)",
		proc.Name, proc.GetPID())
	proc.PID = -1
	return nil
}

func (proc *Process) Stop() error {
	return proc.stop(false)
}

func (proc *Process) ForceStop() error {
	return proc.stop(true)
}

func (proc *Process) IsAlive() error {
	if proc.Proc == nil || proc.Proc.Signal(syscall.Signal(0)) != nil {
		return fmt.Errorf("%s is not started yet", proc.Name)
	}
	return nil
}

func (proc *Process) Restart() error {
	if err := proc.IsAlive(); err == nil {
		if err := proc.Stop(); err != nil {
			return err
		}
		_, _ = proc.WatchState()
	}
	return proc.Start()
}

func (proc *Process) WatchState() (*os.ProcessState, error) {
	if proc.Proc == nil {
		return &os.ProcessState{}, fmt.Errorf("process does not exist")
	}

	return proc.Proc.Wait()
}
