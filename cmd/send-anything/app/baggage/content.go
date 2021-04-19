package baggage

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

// TODO 送る用、受け取る用の構造体に分けるか、統合するかしたい。
type FileInfo struct {
	file     *os.File
	fileName string
	fileDir  string
	size     int64
	hash     []byte
}

func CreateFileInfo(path string) (*FileInfo, error) {
	return createFileInfo(path, false)
}

func createFileInfo(path string, holdDirInfo bool) (*FileInfo, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.Errorf("%v is not file", path)
	}
	if info.IsDir() {
		return nil, errors.Errorf("cannot specify directory: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	d, n := filepath.Split(path)
	if !holdDirInfo {
		// TODO どうするか後で考える
		d = "."
	}

	fInfo := &FileInfo{
		file:     f,
		fileName: n,
		fileDir:  d,
		size:     info.Size(),
		hash:     nil,
	}

	return fInfo, nil
}

func (f *FileInfo) IsExist() bool {
	return f.file != nil
}
func (f *FileInfo) GetChunk(size int) ([]byte, error) {
	chunk := make([]byte, size)
	c, err := f.file.Read(chunk)
	return chunk[:c], err
}
func (f *FileInfo) Name() string {
	return f.fileName
}
func (f *FileInfo) Dir() string {
	return f.fileDir
}

func (f *FileInfo) Path() string {
	return path.Join(f.fileDir, f.fileName)
}

func (f *FileInfo) Size() int64 {
	return f.size
}
func (f *FileInfo) Hash() ([]byte, error) {
	if f.hash != nil && string(f.hash) != "" {
		return f.hash, nil
	}
	f.file.Seek(0, 0)
	h := md5.New()
	if _, err := io.Copy(h, f.file); err != nil {
		return nil, err
	}
	f.hash = h.Sum(nil)
	f.file.Seek(0, 0)
	return f.hash, nil
}

type FilesInfo []*FileInfo

func CreateFilesInfo(filePaths []string) (FilesInfo, error) {
	return createFilesInfo(filePaths, false)
}

func createFilesInfo(filePaths []string, holdDirInfo bool) (FilesInfo, error) {
	fInfo := make([]*FileInfo, 0, len(filePaths))
	errLog := make([]error, 0, len(filePaths))

	for _, path := range filePaths {
		f, err := createFileInfo(path, holdDirInfo)
		if err != nil || f == nil {
			errLog = append(errLog, err)
			continue
		}
		fInfo = append(fInfo, f)
	}

	var err error
	if len(errLog) > 0 {
		err = fmt.Errorf("%v", errLog)
	}
	return fInfo, err
}

func (fs *FilesInfo) IsAllExist() bool {
	for _, f := range *fs {
		if !f.IsExist() {
			return false
		}
	}
	return true
}

func (fs *FilesInfo) Len() int {
	return len(*fs)
}

func CreateFilesInfoByDir(path string) (FilesInfo, error) {
	fPaths, err := getFilePathAllInTree(path)
	if err != nil {
		return nil, err
	}

	filesInfo, err := createFilesInfo(fPaths, true)
	if err != nil {
		return nil, err
	}
	return filesInfo, nil
}

func getFilePathAllInTree(root string, skipDir ...string) ([]string, error) {
	fPaths := make([]string, 0)

	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	trv := filepath.WalkFunc(
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !filepath.IsAbs(path) {
				path, err = filepath.Abs(path)
				if err != nil {
					return err
				}
			}
			if info.IsDir() {
				if isInclude(path, skipDir...) {
					return filepath.SkipDir
				}
				return nil
			} else if info.Mode().IsRegular() {
				fPaths = append(fPaths, path)
			} else {
				return errors.Errorf("unknown path : %v", path)
			}
			return nil
		},
	)
	if err := filepath.Walk(root, trv); err != nil {
		return nil, err
	}
	return fPaths, nil
}

func isInclude(target string, list ...string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func LOG(any interface{}) {
	_, f, l, _ := runtime.Caller(1)
	log.Printf("DEBUG from %s:%d", f, l)
	log.Printf("%v", any)
	log.Printf("")
}
