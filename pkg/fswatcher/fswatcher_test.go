// Copyright (c) 2019-2020 Latona. All rights reserved.
package fswatcher

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"os"
	"reflect"
	"testing"
	"time"
)

var filePath = "test.txt"

func removeFile() {
	_ = os.Remove(filePath)
}

func accessFile(mode int) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|mode, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	fmt.Fprintln(file, "test")
}

func createFile() {
	accessFile(os.O_CREATE)
}

func appendToFile() {
	accessFile(os.O_APPEND)
}

func TestnewWatcher(t *testing.T) {
	testCase := []fsnotify.Op{
		fsnotify.Write,
		fsnotify.Create,
		fsnotify.Remove,
	}

	for fileType := range testCase {
		log.Println(reflect.TypeOf(fileType))
		if _, err := newWatcher([]fsnotify.Op{fsnotify.Op(fileType)}); err != nil {
			t.Error(err)
		}
	}
}

func timeoutSelectCh(fw *FileWatcher) {
	time.Sleep(1 * time.Second)
	fw.StopWatcher()
}

func TestNormal000CreateFileWatcher(t *testing.T) {
	defer removeFile()
	// write file watcher
	fw, err := NewCreateWatcher()
	if err != nil {
		t.Fatal(err)
	}
	if err := fw.AddDirectory("."); err != nil {
		t.Error(err)
	}
	defer fw.StopWatcher()

	createFile()
	go timeoutSelectCh(fw)
	if _, ok := <-fw.GetFilePathCh(); !ok {
		t.Errorf("cant get file path")
	}
	if _, ok := <-fw.GetFilePathCh(); ok {
		t.Errorf("invalid multiple response")
	}
}

func TestAbnormal000CreateFileWatcher(t *testing.T) {
	createFile()
	defer removeFile()
	// write file watcher
	fw, err := NewCreateWatcher()
	if err != nil {
		t.Fatal(err)
	}
	if err := fw.AddDirectory("."); err != nil {
		t.Error(err)
	}
	defer fw.StopWatcher()
	appendToFile()

	go timeoutSelectCh(fw)
	if _, ok := <-fw.GetFilePathCh(); ok {
		t.Errorf("file is remove, but detect append flag")
	}
}

func TestAbnormal001CreateFileWatcher(t *testing.T) {
	createFile()
	defer removeFile()
	// write file watcher
	fw, err := NewCreateWatcher()
	if err != nil {
		t.Fatal(err)
	}
	if err := fw.AddDirectory("."); err != nil {
		t.Error(err)
	}
	defer fw.StopWatcher()
	removeFile()

	go timeoutSelectCh(fw)
	if _, ok := <-fw.GetFilePathCh(); ok {
		t.Errorf("file is remove, but detect append flag")
	}
}

func TestNormal000AppendFileWatcher(t *testing.T) {
	createFile()
	defer removeFile()
	// append to file
	fw, err := NewWriteWatcher()
	if err != nil {
		t.Fatal(err)
	}
	if err := fw.AddDirectory("."); err != nil {
		t.Error(err)
	}
	defer fw.StopWatcher()
	appendToFile()

	go timeoutSelectCh(fw)
	if _, ok := <-fw.GetFilePathCh(); !ok {
		t.Errorf("cant get file path")
	}
	if _, ok := <-fw.GetFilePathCh(); ok {
		t.Errorf("invalid multiple response")
	}
}

func TestAbnormal002AppendFileWatcher(t *testing.T) {
	defer removeFile()
	// append to file
	fw, err := NewWriteWatcher()
	if err != nil {
		t.Fatal(err)
	}
	if err := fw.AddDirectory("."); err != nil {
		t.Error(err)
	}
	defer fw.StopWatcher()
	createFile()

	go timeoutSelectCh(fw)
	if _, ok := <-fw.GetFilePathCh(); ok {
		t.Errorf("cant get file path")
	}
	if _, ok := <-fw.GetFilePathCh(); ok {
		t.Errorf("invalid multiple response")
	}
}

func TestAbnormal001AppendFileWatcher(t *testing.T) {
	defer removeFile()
	// append to file
	fw, err := NewWriteWatcher()
	if err != nil {
		t.Fatal(err)
	}
	createFile()
	if err := fw.AddDirectory("."); err != nil {
		t.Error(err)
	}
	defer fw.StopWatcher()
	removeFile()

	go timeoutSelectCh(fw)
	if _, ok := <-fw.GetFilePathCh(); ok {
		t.Errorf("file is remove, but detect append flag")
	}
}

func TestNormal000RemoveFileWatcher(t *testing.T) {
	defer removeFile()
	createFile()
	//remove file
	fw, err := NewRemoveWatcher()
	if err != nil {
		t.Fatal(err)
	}
	if err = fw.AddDirectory("."); err != nil {
		t.Error(err)
	}
	defer fw.StopWatcher()
	removeFile()
	go timeoutSelectCh(fw)
	if _, ok := <-fw.GetFilePathCh(); !ok {
		t.Errorf("cant get file path")
	}
	if _, ok := <-fw.GetFilePathCh(); ok {
		t.Errorf("invalid multiple response")
	}
}

func TestAbnormal000RemoveFileWatcher(t *testing.T) {
	removeFile()
	// append to file
	fw, err := NewRemoveWatcher()
	if err != nil {
		t.Fatal(err)
	}
	if err := fw.AddDirectory("."); err != nil {
		t.Error(err)
	}
	defer fw.StopWatcher()
	createFile()
	defer removeFile()

	go timeoutSelectCh(fw)
	if _, ok := <-fw.GetFilePathCh(); ok {
		t.Errorf("file is remove, but detect append flag")
	}
}

func TestAbnormal001RemoveFileWatcher(t *testing.T) {
	createFile()
	defer removeFile()
	// append to file
	fw, err := NewRemoveWatcher()
	if err != nil {
		t.Fatal(err)
	}
	if err := fw.AddDirectory("."); err != nil {
		t.Error(err)
	}
	defer fw.StopWatcher()
	appendToFile()

	go timeoutSelectCh(fw)
	if _, ok := <-fw.GetFilePathCh(); ok {
		t.Errorf("file is remove, but detect append flag")
	}
}
