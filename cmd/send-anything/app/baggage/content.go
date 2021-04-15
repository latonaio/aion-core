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

// 後で、送る用、受け取る用の構造体に分けるか、統合するかしたい。
type FileInfo struct {
	file     *os.File
	fileName string
	fileDir  string
	size     int64
	hash     []byte
}

func CreateFileInfo(path string) (*FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.Errorf("cannot specify directory: %s", path)
	}
	d, n := filepath.Split(path)
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
func (f *FileInfo) GetChunk(size []byte) (int, error) {
	// 後で、バイトサイズ指定し、そのサイズの[]byteを返すように修正するか。
	return f.file.Read(size)
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
	fInfo := make([]*FileInfo, 0, len(filePaths))
	errLog := make([]error, 0, len(filePaths))

	for _, path := range filePaths {
		f, err := CreateFileInfo(path)
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
func LOG(any interface{}) {
	_, f, l, _ := runtime.Caller(1)
	log.Printf("DEBUG from %s:%d", f, l)
	log.Printf("%v", any)
	log.Printf("")
}

type DirInfo struct {
	dirPath   string
	files     *FilesInfo
	subDirs   *[]*DirInfo
	parentDir *DirInfo
}

func CreateDirInfo(path string) (*DirInfo, error) {

	return nil, nil
}
