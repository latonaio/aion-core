package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"bitbucket.org/latonaio/aion-core/cmd/kanban-server/app"
	"bitbucket.org/latonaio/aion-core/cmd/send-anything/app/baggage"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"github.com/avast/retry-go"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
)

func sendToServerNotifyFailure(stream kanbanpb.SendAnything_SendToOtherDevicesClient) error {
	req := &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_SendingFile_FAILED,
		Context: nil,
	}
	if err := sendToServer(stream, req); err != nil {
		return xerrors.Errorf("場所 : %w", err)
	}
	return nil
}

func sendToServer(stream kanbanpb.SendAnything_SendToOtherDevicesClient, req *kanbanpb.SendContext) error {
	const connRetryCount = 10
	doFunc := func() error {
		return stream.Send(req)
	}
	delayFunc := func(n uint, config *retry.Config) time.Duration {
		log.Warnf("[SendToOtherDevices client] retry to send because failed to send buffer ")
		return time.Second
	}

	if err := retry.Do(doFunc, retry.DelayType(delayFunc), retry.Attempts(connRetryCount)); err != nil {
		log.Errorf("[SendToOtherDevices client] failed to send chunk(error: %v)", err)
		return xerrors.Errorf("場所 : %w", err)
	}
	return nil
}

func sendToOtherDeviceClient(ctx context.Context, m *kanbanpb.SendKanban, port int) (err error) {
	// connect to remote device
	deviceAddr := m.DeviceAddr + ":" + strconv.Itoa(port)
	conn, err := grpc.DialContext(ctx, deviceAddr, grpc.WithInsecure())
	if err != nil {
		return xerrors.Errorf("cannot connect to remote device: %s: %w", deviceAddr, err)
	}
	defer conn.Close()

	log.Printf("[SendToOtherDevice client] success to connect : %s", deviceAddr)

	client := kanbanpb.NewSendAnythingClient(conn)
	stream, err := client.SendToOtherDevices(ctx)
	if err != nil {
		return xerrors.Errorf("cannot connect to remote device: %s: %w", deviceAddr, err)
	}

	// close stream
	defer func() {
		reply, err := stream.CloseAndRecv()
		if err != nil {
			log.Printf("[SendToOtherDevice client] received eos ack is failed: %s: %+v", deviceAddr, err)
			return
		}
		if reply.StatusCode != kanbanpb.UploadStatusCode_OK {
			log.Printf("[SendToOtherDevice client] response error message: %s", reply.Message)
			return
		}
		log.Printf("[SendToOtherDevice client] success to send kanban and files: %s", deviceAddr)
	}()

	// send kanban
	if err := sendKanban(stream, m); err != nil {
		return xerrors.Errorf("[SendToOtherDevice client] sending kanban is failed destination　address: %s: %w", deviceAddr, err)
	}

	// eos kanban
	// ファイル送信関数でpanicになっても実行させる
	sendFileCnt := 0
	defer func() {
		e := sendEOS(stream, sendFileCnt)
		if e != nil {
			if err != nil {
				err = xerrors.Errorf("[SendToOtherDevice client] sending kanban is failed destination　address: %s: %w", deviceAddr, e)
			}
		}
	}()

	if filepaths := m.AfterKanban.FileList; len(filepaths) > 0 {
		if sendFileCnt, err = sendAnything(stream, filepaths); err != nil {
			sendToServerNotifyFailure(stream)
			LOG(fmt.Sprintf("%+v\n", err))
			return xerrors.Errorf("destination　address: %s: %w", deviceAddr, err)
		}
	}

	return nil
}

// send kanban to other devices
func sendKanban(stream kanbanpb.SendAnything_SendToOtherDevicesClient, m *kanbanpb.SendKanban) error {
	log.Printf("[SendToOtherDevices client] start to send kanban: %s", m.NextService)
	cont, err := anypb.New(m)
	if err != nil {
		return xerrors.Errorf("cannot create kanban: %w", err)
	}

	req := &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_SendingKanban,
		Context: cont,
	}
	if err := stream.Send(req); err != nil {
		return xerrors.Errorf("failed sending kanban: %w", err)
	}
	log.Printf("[SendToOtherDevices client] finish to send kanban: %s", m.NextService)
	return nil
}

func sendAnything(stream kanbanpb.SendAnything_SendToOtherDevicesClient, paths []string) (int, error) {
	env := app.GetConfig()
	sendableDirRoot := filepath.Clean(env.AionHome)
	itemInfo := newSentItemInfo()

	for _, p := range paths {
		p = getAbsPath(sendableDirRoot, p)
		pathInfo, err := os.Stat(p)
		if err != nil {
			return -1, xerrors.Errorf("failed getting stat: %w", err)
		}
		switch m := pathInfo.Mode(); {
		case m.IsDir():
			d, err := sendDir(stream, p, sendableDirRoot)
			if err != nil {
				return -1, xerrors.Errorf("failed sending directory: %w", err)
			}

			itemInfo.dirs = append(itemInfo.dirs, d)

		case m.IsRegular():
			f, err := sendFiles(stream, []string{p}, sendableDirRoot)
			if err != nil {
				return -1, xerrors.Errorf("failed sending files: %w", err)
			}

			itemInfo.files = append(itemInfo.files, f...)
		}
	}
	return itemInfo.sentFileCnt(), nil
}

type sentItemInfo struct {
	files []*sentFileInfo
	dirs  []*sentDirInfo
}

type sentDirInfo struct {
	fileCnt          int
	dirName          string
	relParentDirPath string
	files            []*sentFileInfo
}

type sentFileInfo struct {
	dirName    string
	relDirPath string
}

func newSentItemInfo() *sentItemInfo {
	return &sentItemInfo{
		// make関数: スライス作成時、容量(cap)は長さと同じになる
		files: make([]*sentFileInfo, 0),
		dirs:  make([]*sentDirInfo, 0),
	}
}

func (i *sentItemInfo) sentFileCnt() int {
	cnt := 0
	for _, d := range i.dirs {
		cnt += d.fileCnt
	}
	cnt += len(i.files)
	return cnt
}

func sendDir(stream kanbanpb.SendAnything_SendToOtherDevicesClient, dirPath, aionHome string) (*sentDirInfo, error) {
	files, err := takeFiles(aionHome, []string{dirPath})
	if err != nil {
		return nil, xerrors.Errorf("cannot take in files: %w", err)
	}

	// directory送信で、一番最初に送るやつ(の材料)
	dirInfo := &sentDirInfo{
		fileCnt:          0,
		dirName:          filepath.Base(dirPath),
		relParentDirPath: ".", // 送信ディレクトリの親Dir情報は破棄する。今後の改修に期待……
		files:            make([]*sentFileInfo, 0, len(files)),
	}

	if !filepath.IsAbs(dirPath) {
		dirPath = filepath.Join(aionHome, dirPath)
	}

	err = sendDirInfo(stream, dirInfo, len(files))
	if err != nil {
		return nil, xerrors.Errorf("failed sending directory info: %w", err)
	}

	for _, f := range files {
		relPath, err := filepath.Rel(dirPath, f.Path())
		if err != nil {
			return nil, xerrors.Errorf("cannot find relational path %s to %s: %w", dirPath, f.Path(), err)
		}
		info, err := sendFile(stream, f, filepath.Dir(relPath))
		if err != nil {
			return nil, xerrors.Errorf("failed sending file: %w", err)
		}
		dirInfo.fileCnt++
		dirInfo.files = append(dirInfo.files, info)
	}

	err = sendEndOfDir(stream, dirInfo)
	if err != nil {
		return nil, xerrors.Errorf("failed sending EndOfDir: %w", err)
	}
	return dirInfo, nil
}

func sendDirInfo(stream kanbanpb.SendAnything_SendToOtherDevicesClient, dirInfo *sentDirInfo, fileCnt int) error {
	return createDirInfo(stream, dirInfo, fileCnt, kanbanpb.UploadRequestCode_SendingDirInfo)
}

func sendEndOfDir(stream kanbanpb.SendAnything_SendToOtherDevicesClient, dirInfo *sentDirInfo) error {
	return createDirInfo(stream, dirInfo, dirInfo.fileCnt, kanbanpb.UploadRequestCode_EndOfSendingDir)
}

// TODO 名前がよくない
func createDirInfo(stream kanbanpb.SendAnything_SendToOtherDevicesClient, dirInfo *sentDirInfo, fileCnt int, code kanbanpb.UploadRequestCode) error {
	context, err := anypb.New(&kanbanpb.DirInfo{
		FileCnt:       int64(fileCnt),
		DirName:       dirInfo.dirName,
		RelParentPath: dirInfo.relParentDirPath,
	})
	if err != nil {
		return xerrors.Errorf("cannot create kanban: %w", err)
	}

	err = sendToServer(stream, &kanbanpb.SendContext{
		Code:    code,
		Context: context,
	})
	if err != nil {
		return xerrors.Errorf("failed sending request: %w", err)
	}
	return nil
}

// filePaths 絶対パスでも、aionHomeからの相対パスでも、どっちでも良い
func sendFiles(stream kanbanpb.SendAnything_SendToOtherDevicesClient, filePaths []string, aionHome string) ([]*sentFileInfo, error) {
	files, err := takeFiles(aionHome, filePaths)
	if err != nil {
		return nil, xerrors.Errorf("cannot take in files: %w", err)
	}

	sentFileInfoList := make([]*sentFileInfo, 0, len(files))

	for _, f := range files {
		relPath := f.Path()
		if filepath.IsAbs(relPath) {
			relPath, err = filepath.Rel(aionHome, f.Path())
			if err != nil {
				return nil, xerrors.Errorf("cannot find relational path %s to %s: %w", aionHome, f.Path(), err)
			}
		}

		info, err := sendFile(stream, f, filepath.Dir(relPath))
		if err != nil {
			return nil, xerrors.Errorf("failed sending file: %w", err)
		}
		sentFileInfoList = append(sentFileInfoList, info)
	}

	return sentFileInfoList, nil
}

// pathsList[.]が絶対パスの場合はそれを。
// 相対パスの場合は、relPathFrom/pathsList[.] のファイルを探す。
func takeFiles(relPathFrom string, pathsList ...[]string) (baggage.FilesInfo, error) {
	files := make(baggage.FilesInfo, 0, len(pathsList))

	for _, paths := range pathsList {
		for _, path := range paths {
			path = getAbsPath(relPathFrom, path)
			if !isSendable(path, relPathFrom) {
				return nil, xerrors.Errorf("file %v is not allowd sending", path)
			}

			info, err := os.Stat(path)
			if err != nil {
				return nil, xerrors.Errorf("failed getting stat: %w", err)
			}

			switch m := info.Mode(); {
			case m.IsDir():
				fs, err := baggage.CreateFilesInfoByDir(path)
				if err != nil {
					return nil, xerrors.Errorf("failed create files info by %s: %w", path, err)
				}
				files = append(files, fs...)
			case m.IsRegular():
				fs, err := baggage.CreateFileInfo(path)
				if err != nil {
					return nil, xerrors.Errorf("failed create files info by %s: %w", err)
				}
				files = append(files, fs)
			default:
				return nil, xerrors.New("unknown FileSystem type")
			}
		}
	}

	return files, nil
}

// send data to other devices

// relDir AionHomeまたは送信directoryから、送信ファイルのカレントディレクトリまでの相対パス
func sendFile(stream kanbanpb.SendAnything_SendToOtherDevicesClient, file *baggage.FileInfo, relDir string) (*sentFileInfo, error) {
	if file == nil || !file.IsExist() {
		return nil, xerrors.New("skip sending a file")
	}
	log.Printf("[SendToOtherDevices client] start to send file: %s", file.Name())

	err := sendFileInfo(stream, file, relDir)
	if err != nil {
		return nil, xerrors.Errorf("failed sending file info: %w", err)
	}

	chunkCnt, err := sendFileContent(stream, file)
	if err != nil {
		return nil, xerrors.Errorf("failed sending file content: %w", err)
	}

	err = sendEOF(stream, file, relDir, chunkCnt)
	if err != nil {
		return nil, xerrors.Errorf("failed sending EndOfFile: %w", err)
	}

	log.Printf("[SendToOtherDevices client] finish to send file: %s", file.Name())
	return &sentFileInfo{
		dirName:    file.Name(),
		relDirPath: relDir,
	}, nil
}

// relDir AionHomeから、またはdirectoryから、送信ファイルのカレントディレクトリまでの相対パス
func sendFileInfo(stream kanbanpb.SendAnything_SendToOtherDevicesClient, f *baggage.FileInfo, relDir string) error {
	return createFileInfo(stream, f, relDir, int32(f.Size()/chunkSize), kanbanpb.UploadRequestCode_SendingFile_Info)
}

// relDir AionHome、またはdirectoryから、送信ファイルのカレントディレクトリまでの相対パス
func sendEOF(stream kanbanpb.SendAnything_SendToOtherDevicesClient, f *baggage.FileInfo, relDir string, chunkCnt int32) error {
	return createFileInfo(stream, f, relDir, chunkCnt, kanbanpb.UploadRequestCode_SendingFile_EOF)
}

// TODO 名前が悪い
func createFileInfo(stream kanbanpb.SendAnything_SendToOtherDevicesClient, f *baggage.FileInfo, relDir string, chunkCnt int32, code kanbanpb.UploadRequestCode) error {
	hash, err := f.Hash()
	if err != nil {
		return xerrors.Errorf("cananot get file hash: %w", err)
	}

	fInfo, err := anypb.New(&kanbanpb.FileInfo{
		Size:     f.Size(),
		ChunkCnt: chunkCnt,
		Hash:     hash,
		Name:     f.Name(),
		RelDir:   relDir,
	})
	if err != nil {
		return xerrors.Errorf("cannot create kanban: %w", err)
	}

	err = sendToServer(stream, &kanbanpb.SendContext{
		Code:    code,
		Context: fInfo,
	})
	if err != nil {
		return xerrors.Errorf("failed sending request: %w", err)
	}
	return nil
}

func sendFileContent(stream kanbanpb.SendAnything_SendToOtherDevicesClient, f *baggage.FileInfo) (int32, error) {
	sendChunkCnt := int32(0)
	for ; true; sendChunkCnt++ {
		chunk, err := f.GetChunk(chunkSize)
		if err == io.EOF {
			break
		} else if err != nil {
			return -1, xerrors.Errorf("failed getting file chunk. No %d: %w", sendChunkCnt, err)
		}

		err = sendFileChunk(stream, chunk, f.Name(), sendChunkCnt)
		if err != nil {
			return -1, xerrors.Errorf("failed sending file chunk. No %d: %w", sendChunkCnt, err)
		}
	}
	return sendChunkCnt, nil
}

func sendFileChunk(stream kanbanpb.SendAnything_SendToOtherDevicesClient, chunk []byte, fName string, refNum int32) error {
	req, err := createRequestChunkFile(chunk, fName, refNum)
	if err != nil {
		return xerrors.Errorf("cannot create chunk kanban: %w", err)
	}

	if err := sendToServer(stream, req); err != nil {
		return xerrors.Errorf("failed sending request: %w", err)
	}
	return nil
}

func createRequestChunkFile(chunk []byte, name string, refNum int32) (*kanbanpb.SendContext, error) {
	anyChunk, err := anypb.New(&kanbanpb.Chunk{
		Context: chunk,
		Name:    name,
		RefNum:  refNum,
	})
	if err != nil {
		return nil, xerrors.Errorf("cannot create kanban: %w", err)
	}

	return &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_SendingFile_CONT,
		Context: anyChunk,
	}, nil
}

// send eos to other devices
func sendEOS(stream kanbanpb.SendAnything_SendToOtherDevicesClient, fileCnt int) (err error) {
	eos, err := anypb.New(&kanbanpb.StreamInfo{
		FileCount: int32(fileCnt),
	})

	// streamInfoの作成に失敗しても、とりあえずEOSは送りたい。
	// sendのエラーがnilなら、kanban作成時のエラーを返す。
	// sendが失敗したら、sendのエラーを優先的に返す。
	if err != nil {
		defer func(e error) {
			if err == nil {
				err = e
			}
		}(xerrors.Errorf("cannot create kanban: %w", err))
	}
	req := &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_EOS,
		Context: eos,
	}
	if err := stream.Send(req); err != nil {
		return xerrors.Errorf("failed sending request: %w", err)
	}
	return nil
}

func LOG(any ...interface{}) {
	// log.Println(any...)
	// return
	_, f, l, _ := runtime.Caller(1)
	log.Printf("DEBUG from %s:%d", f, l)

	if len(any) > 2 {
		format, ok := any[0].(string)
		if ok {
			f := fmt.Sprintf(format, any[1:]...)
			ln := fmt.Sprintln(any...)
			if len(f) > len(ln) {
				log.Println(ln)
			} else {
				log.Println(f)
			}
			log.Println()
			return
		}
	}

	log.Println(any...)
	log.Println()
}

func isSendable(target, relPathFrom string) bool {
	var err error
	if err == nil {
		target, err = filepath.Abs(target)
	}
	if err == nil {
		relPathFrom, err = filepath.Abs(relPathFrom)
	}

	return err == nil && filepath.HasPrefix(target, relPathFrom)
}
func getAbsPath(root, relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}
	return filepath.Join(root, relPath)
}
