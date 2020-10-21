// Copyright (c) 2019-2020 Latona. All rights reserved.
package fswatcher

import (
	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher    *fsnotify.Watcher
	filePathCh chan string
}

func newWatcher(fileType []fsnotify.Op) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		watcher:    watcher,
		filePathCh: make(chan string, 1),
	}

	go func() {
		for event := range watcher.Events {
			for _, ft := range fileType {
				if event.Op&fsnotify.Op(ft) == fsnotify.Op(ft) {
					fw.filePathCh <- event.Name
				}
			}
		}

		close(fw.filePathCh)
	}()

	return fw, nil
}

func NewWriteWatcher() (*FileWatcher, error) {
	return newWatcher([]fsnotify.Op{fsnotify.Write})
}

func NewCreateWatcher() (*FileWatcher, error) {
	return newWatcher([]fsnotify.Op{fsnotify.Create})
}

func NewRemoveWatcher() (*FileWatcher, error) {
	return newWatcher([]fsnotify.Op{fsnotify.Remove})
}

func NewWriteAndRemoveWatcher() (*FileWatcher, error) {
	return newWatcher([]fsnotify.Op{fsnotify.Write, fsnotify.Remove})
}

func (fw *FileWatcher) GetFilePathCh() chan string {
	return fw.filePathCh
}

func (fw *FileWatcher) AddDirectory(dirPath string) error {
	return fw.watcher.Add(dirPath)
}

func (fw *FileWatcher) RemoveDirectory(dirPath string) error {
	return fw.watcher.Remove(dirPath)
}

func (fw *FileWatcher) StopWatcher() {
	fw.watcher.Close()
}
