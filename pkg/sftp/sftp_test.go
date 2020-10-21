// Copyright (c) 2019-2020 Latona. All rights reserved.
package sftp

import (
	"bitbucket.org/latonaio/aion-core/pkg/fswatcher"
	"os"
	"path"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

const (
	testDataDir  = "../../test/test_data/sftp"
	testFileName = "test.json"
)

var (
	srcDir         = path.Join(testDataDir, "send")
	dstDir         = path.Join(testDataDir, "receive")
	srcFilePath    = path.Join(srcDir, testFileName)
	dstFilePath, _ = filepath.Abs(path.Join(dstDir, testFileName))
)

type testCase struct {
	user string
	pass string
	host string
	port int
}

type sendFileContainer struct {
	src string
	dst string
}

var normalCase = []testCase{
	{"latona", "Latona2019", "localhost", 22},
}

var abnormalCase = []testCase{
	{"aa", "Latona2019", "localhost", 22},
	{"latona", "aa", "localhost", 22},
	{"latona", "Latona2019", "aa", 22},
	{"latona", "Latona2019", "localhost", 11},
}

var sendFileNormalCase = []sendFileContainer{
	{srcFilePath, dstFilePath},
}

var sendFileAbnormalCase = []sendFileContainer{
	{dstFilePath, srcFilePath},
	{srcFilePath, "abcdefg/abcdefg"},
}

func removeTestFile(filePath string, t *testing.T) {
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNormalCaseNewSFTPClient(t *testing.T) {
	for _, tc := range normalCase {
		t.Run("normalcase", func(t *testing.T) {
			sc := NewSFTPClient(tc.user, tc.pass, tc.host, tc.port)
			defer sc.Close()
		})
	}
}

// multiple connection
func TestNormalCase002NewSFTPClient(t *testing.T) {
	for _, tc := range normalCase {
		var wg sync.WaitGroup
		t.Run("normalcase: multiple connection", func(t *testing.T) {
			for range make([]int, 5) {
				wg.Add(1)
				go func() {
					defer wg.Done()
					sc := NewSFTPClient(tc.user, tc.pass, tc.host, tc.port)
					defer sc.Close()
				}()
			}
			wg.Wait()
		})
	}
}

func TestNormalCaseSendFile(t *testing.T) {
	tc := normalCase[0]
	ss := NewSFTPClient(tc.user, tc.pass, tc.host, tc.port)
	defer ss.Close()

	for _, files := range sendFileNormalCase {
		t.Run("normalcase", func(t *testing.T) {
			removeTestFile(files.dst, t)
			defer removeTestFile(files.dst, t)

			watcher, _ := fswatcher.NewCreateWatcher()
			_ = watcher.AddDirectory(path.Dir(files.dst))

			if err := ss.SetToSendPathCh(files.src, files.dst); err != nil {
				t.Error(err)
			}

			timeout := make(chan bool, 1)
			go func() {
				time.Sleep(1 * time.Second)
				timeout <- true
			}()

			select {
			case <-watcher.GetFilePathCh():
				break
			case <-timeout:
				t.Errorf("cant create file : %s", files.dst)
			}
		})
	}
}

func TestAbnormalCase001SendFile(t *testing.T) {
	tc := normalCase[0]
	ss := NewSFTPClient(tc.user, tc.pass, tc.host, tc.port)
	defer ss.Close()

	for _, files := range sendFileAbnormalCase {
		t.Run("abnormalcase", func(t *testing.T) {
			watcher, _ := fswatcher.NewCreateWatcher()
			_ = watcher.AddDirectory(path.Dir(files.dst))

			if err := ss.SetToSendPathCh(files.src, files.dst); err != nil {
				t.Error(err)
			}

			timeout := make(chan bool, 1)
			go func() {
				time.Sleep(1 * time.Second)
				timeout <- true
			}()

			select {
			case <-watcher.GetFilePathCh():
				t.Errorf("set invalid param, but can create file")
			case <-timeout:
				break
			}
		})
	}
}

func TestAbnormalCase002SendFile(t *testing.T) {
	tc := normalCase[0]
	ss := NewSFTPClient(tc.user, tc.pass, tc.host, tc.port)
	defer ss.Close()

	files := sendFileNormalCase[0]
	watcher, _ := fswatcher.NewCreateWatcher()
	_ = watcher.AddDirectory(path.Dir(files.dst))

	ss.Close()
	if err := ss.SetToSendPathCh(files.src, files.dst); err != nil {
		t.Error(err)
	}

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()

	select {
	case <-watcher.GetFilePathCh():
		t.Errorf("set invalid param, but can create file")
	case <-timeout:
		break
	}
}
