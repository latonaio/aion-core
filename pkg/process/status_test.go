// Copyright (c) 2019-2020 Latona. All rights reserved.
package process

import (
	"testing"
)

func TestAddWatcher(t *testing.T) {
	watcher := NewWatcher()

	proc, err := NewProcess("test", "test ls", defaultDirPath, defaultEnv)
	if err != nil {
		t.Fatal(err)
	}

	watcher.Add(proc)
	if len(watcher.processStatusList) != 1 {
		t.Errorf("cant set process to watcher")
	}
}

func TestStopProcess(t *testing.T) {
	watcher := NewWatcher()

	// normal case
	proc, err := NewProcess("test", "test ls", defaultDirPath, defaultEnv)
	if err != nil {
		t.Fatal(err)
	}
	watcher.Add(proc)

	// normal case
	if err := watcher.Stop(proc.GetPID()); err != nil {
		t.Error(err)
	}
	// abnormal case
	if err := watcher.Stop(1); err == nil {
		t.Errorf("pid 1 is not watched, but stop is success")
	}
}

func TestGetExitProcessCh(t *testing.T) {
	watcher := NewWatcher()

	// normal case
	proc, err := NewProcess("test", "test ls", defaultDirPath, defaultEnv)
	if err != nil {
		t.Fatal(err)
	}
	watcher.Add(proc)
	exitProcessCh := watcher.GetExitProcessCh()

	if _, ok := <-exitProcessCh; !ok {
		t.Errorf("cant get exit status")
	}

}
