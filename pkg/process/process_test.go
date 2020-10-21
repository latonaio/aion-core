// Copyright (c) 2019-2020 Latona. All rights reserved.
package process

import (
	"testing"
)

type testCase struct {
	name    string
	command string
}

const defaultDirPath = "."

var defaultEnv []string

func TestNewProcess(t *testing.T) {
	testCases := []testCase{
		{"test1", "test"},
		{"test2", "test ls"},
	}
	for _, tc := range testCases {
		proc, err := NewProcess(tc.name, tc.command, defaultDirPath, defaultEnv)
		if err != nil {
			t.Fatal(err)
		}
		if proc.PID == 0 {
			t.Errorf("process is null : %d", proc.PID)
		}
	}

	abnormalTestCases := []testCase{
		{"test3", "aaaa"},
	}
	for _, tc := range abnormalTestCases {
		_, err := NewProcess(tc.name, tc.command, defaultDirPath, defaultEnv)
		if err == nil {
			t.Fatal(err)
		}
	}
}

func TestIsAlive(t *testing.T) {
	testCases := []testCase{
		{"test1", "sleep 5"},
	}
	for _, tc := range testCases {
		proc, err := NewProcess(tc.name, tc.command, defaultDirPath, defaultEnv)
		if err != nil {
			t.Fatal(err)
		}
		if err = proc.IsAlive(); err != nil {
			t.Fatal(err)
		} else {
			proc.Proc.Kill()
		}
	}

	abnormalTestCases := []testCase{
		{"test1", "test ls"},
	}
	for _, tc := range abnormalTestCases {
		proc, err := NewProcess(tc.name, tc.command, defaultDirPath, defaultEnv)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = proc.Proc.Wait()
		if err := proc.IsAlive(); err == nil {
			t.Errorf("cant get exit status")
		}
	}
}

func TestStart(t *testing.T) {
	// normal case
	proc, err := NewProcess(
		"test", "test ls", defaultDirPath, defaultEnv)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = proc.Proc.Wait()
	if err := proc.Start(); err != nil {
		t.Error(err)
	}
	// abnormal case
	if err := proc.Start(); err == nil {
		t.Errorf("proc is already started, but cant detect it")
	}
	// abnormal case
	abnormalProc := Process{
		Cmd:  "aaa",
		Args: []string{},
	}
	if err := abnormalProc.Start(); err == nil {
		t.Errorf("cant detect to invalid command")
	}
}

func TestStop(t *testing.T) {
	// normal case
	proc, err := NewProcess("test", "sleep 5", defaultDirPath, defaultEnv)
	if err != nil {
		t.Fatal(err)
	}
	if err := proc.Stop(); err != nil {
		t.Error(err)
	}
	_, _ = proc.Proc.Wait()
	// abnormal case
	if err := proc.Stop(); err == nil {
		t.Errorf("process already exist, but execute to stop")
	}
}

func TestForceStop(t *testing.T) {
	// normal case
	proc, err := NewProcess("test", "sleep 5", defaultDirPath, defaultEnv)
	if err != nil {
		t.Fatal(err)
	}
	if err := proc.ForceStop(); err != nil {
		t.Error(err)
	}
	// abnormal case
	if err := proc.ForceStop(); err == nil {
		t.Errorf("process already exist, but execute to stop")
	}
}

func TestWatchState(t *testing.T) {
	proc, err := NewProcess("test", "test ls", defaultDirPath, defaultEnv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := proc.WatchState(); err != nil {
		t.Error(err)
	}
	// abnormal case
	if _, err := proc.WatchState(); err == nil {
		t.Errorf("process already exist, but execute to watchState")
	}
	// abnormal case
	abnormalProc := Process{
		Cmd:  "aaa",
		Args: []string{},
	}
	if _, err := abnormalProc.WatchState(); err == nil {
		t.Errorf("proc.Proc is nil, but success")
	}
}

func TestRestart(t *testing.T) {
	proc, err := NewProcess("test", "test ls", defaultDirPath, defaultEnv)
	if err != nil {
		t.Fatal(err)
	}
	if err := proc.Restart(); err != nil {
		t.Error(err)
	}
	_, _ = proc.Proc.Wait()
	if err := proc.Restart(); err != nil {
		t.Error(err)
	}
	// abnormal case
	abnormalProc := Process{
		Cmd:  "aaa",
		Args: []string{},
	}
	if err := abnormalProc.Restart(); err == nil {
		t.Errorf("cant stop process, but can restart")
	}
}

func TestGetPID(t *testing.T) {
	proc, err := NewProcess("test", "test ls", defaultDirPath, defaultEnv)
	if err != nil {
		t.Fatal(err)
	}
	if pid := proc.GetPID(); pid == 0 {
		t.Errorf("cant get pid")
	}
}
