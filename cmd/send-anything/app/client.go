package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/latonaio/aion-core/cmd/kanban-server/app"
	"bitbucket.org/latonaio/aion-core/cmd/send-anything/app/baggage"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
)

func sendToServerNotifyFailure(stream kanbanpb.SendAnything_SendToOtherDevicesClient) error {
	req := &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_SendingFile_FAILED,
		Context: nil,
	}
	if err := sendToServer(stream, req); err != nil {
		return err
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
		return err
	}
	return nil
}

func sendToOtherDeviceClient(ctx context.Context, m *kanbanpb.SendKanban, port int) (err error) {
	// connect to remote device
	deviceAddr := m.DeviceAddr + ":" + strconv.Itoa(port)
	conn, err := grpc.DialContext(ctx, deviceAddr, grpc.WithInsecure())
	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("cannot connect to remote device: %s", deviceAddr))
		log.Print(err)
		return err
	}
	defer conn.Close()

	log.Printf("[SendToOtherDevice client] success to connect : %s", deviceAddr)

	client := kanbanpb.NewSendAnythingClient(conn)
	stream, err := client.SendToOtherDevices(ctx)
	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("cannot connect to remote device: %s", deviceAddr))
		log.Print(err)
		return err
	}

	// close stream
	defer func() {
		reply, err := stream.CloseAndRecv()
		if err != nil {
			log.Printf("[SendToOtherDevice client] received eos ack is failed: %s: %v", deviceAddr, err)
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
		return errors.Wrap(err, fmt.Sprintf("[SendToOtherDevice client] sending kanban is failed destination　address: %s", deviceAddr))
	}

	// eos kanban
	// ファイル送信関数でpanicになっても実行される。
	sendFileCnt := 0
	defer func() {
		e := sendEOS(stream, sendFileCnt)
		if e != nil {
			if err != nil {
				err = errors.Wrap(e, fmt.Sprintf("[SendToOtherDevice client] sending kanban is failed destination　address: %s", deviceAddr))
			}
		}
	}()

	if filepaths := m.AfterKanban.FileList; len(filepaths) > 0 {
		if sendFileCnt, err = sendFiles(stream, filepaths); err != nil {
			sendToServerNotifyFailure(stream)
			return errors.Wrap(err, fmt.Sprintf("destination　address: %s", deviceAddr))
		}
	}

	return nil
}

// send kanban to other devices
func sendKanban(stream kanbanpb.SendAnything_SendToOtherDevicesClient, m *kanbanpb.SendKanban) error {
	log.Printf("[SendToOtherDevices client] start to send kanban: %s", m.NextService)
	cont, err := anypb.New(m)
	if err != nil {
		return err
	}

	req := &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_SendingKanban,
		Context: cont,
	}
	if err := stream.Send(req); err != nil {
		return err
	}
	log.Printf("[SendToOtherDevices client] finish to send kanban: %s", m.NextService)
	return nil
}

func sendFiles(stream kanbanpb.SendAnything_SendToOtherDevicesClient, filePaths []string) (int, error) {
	env := app.GetConfig()
	sendableDirRoot := filepath.Clean(env.AionHome)
	files, err := takeFiles(sendableDirRoot, filePaths)
	if err != nil {
		return -1, err
	}
	// 全て送るか、全て送らないかにしている。後で、選択できるようにしたい。
	if !files.IsAllExist() {
		invalidFilePaths := make([]string, 0, len(files))
		for i, f := range files {
			if !f.IsExist() {
				invalidFilePaths = append(invalidFilePaths, filePaths[i])
			}
		}
		return -1, errors.Errorf("cannot open file(s) %v", invalidFilePaths)
	}

	errStr := make([]string, 0, len(files))
	for _, f := range files {
		if err := sendFile(stream, f, env.AionHome); err != nil {
			errStr = append(errStr, err.Error())
		}
	}
	if len(errStr) != 0 {
		return (files.Len() - len(errStr)), fmt.Errorf("cant send files\n%s", strings.Join(errStr, "\n"))
	}
	return files.Len(), nil
}

func takeFiles(relPathFrom string, pathsList ...[]string) (baggage.FilesInfo, error) {
	files := make(baggage.FilesInfo, 0, len(pathsList))
	LOG(pathsList)

	for _, paths := range pathsList {
		for _, path := range paths {
			path = getAbsPath(relPathFrom, path)
			if !isSendable(path, relPathFrom) {
				return nil, errors.Errorf("file %v is not allowd sending", path)
			}

			info, err := os.Stat(path)
			if err != nil {
				LOG(err)
				return nil, err
			}

			LOG("ok")
			switch m := info.Mode(); {
			case m.IsDir():
				LOG("IS DIR")
				fs, err := baggage.CreateFilesInfoByDir(path)
				if err != nil {
					LOG(err)
					return nil, err
				}
				files = append(files, fs...)
				LOG("ok")
			case m.IsRegular():
				LOG("IS FILE")
				fs, err := baggage.CreateFileInfo(path)
				if err != nil {
					return nil, err
				}
				files = append(files, fs)
				LOG("ok")
			default:
				return nil, fmt.Errorf("unknown FileSystem type")
			}
		}
	}

	return files, nil
}

// send data to other devices
func sendFile(stream kanbanpb.SendAnything_SendToOtherDevicesClient, file *baggage.FileInfo, parentDir string) error {
	if file == nil || !file.IsExist() {
		return errors.Errorf("skip sending a file")
	}
	log.Printf("[SendToOtherDevices client] start to send file: %s", file.Name())

	err := sendFileInfo(stream, file, parentDir)
	if err != nil {
		return err
	}
	LOG("FIN SEND FILE INFO")

	chunkCnt, err := sendFileContent(stream, file)
	if err != nil {
		return err
	}

	LOG("FIN SEND FILE CONTENT")

	err = sendFileEOF(stream, file, parentDir, chunkCnt)
	if err != nil {
		return err
	}

	LOG("FIN SEND FILE EOF")

	log.Printf("[SendToOtherDevices client] finish to send file: %s", file.Name())
	return nil
}

func sendFileInfo(stream kanbanpb.SendAnything_SendToOtherDevicesClient, f *baggage.FileInfo, usableDir string) error {
	hash, err := f.Hash()
	if err != nil {
		return err
	}

	relDir := f.Dir()
	if filepath.IsAbs(relDir) {
		relDir, err = filepath.Rel(usableDir, f.Dir())
		if err != nil {
			return err
		}
	}

	anyInfo, err := anypb.New(
		&kanbanpb.FileInfo{
			Size:     f.Size(),
			Hash:     hash,
			Name:     f.Name(),
			ChunkCnt: int32(f.Size() / chunkSize),
			RelDir:   relDir,
		},
	)
	if err != nil {
		return err
	}

	err = sendToServer(stream, &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_SendingFile_Info,
		Context: anyInfo,
	})
	if err != nil {
		return err
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
			return -1, err
		}

		err = sendFileChunk(stream, chunk, f.Name(), sendChunkCnt)
		if err != nil {
			return -1, err
		}
	}
	return sendChunkCnt, nil
}

func sendFileChunk(stream kanbanpb.SendAnything_SendToOtherDevicesClient, chunk []byte, fName string, refNum int32) error {
	req, err := createRequestChunkFile(chunk, fName, refNum)
	if err != nil {
		return err
	}

	if err := sendToServer(stream, req); err != nil {
		return err
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
		return nil, err
	}

	return &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_SendingFile_CONT,
		Context: anyChunk,
	}, nil
}

func sendFileEOF(stream kanbanpb.SendAnything_SendToOtherDevicesClient, f *baggage.FileInfo, sendableDirRoot string, chunkCnt int32) error {
	relDir := f.Dir()
	var err error
	if filepath.IsAbs(relDir) {
		relDir, err = filepath.Rel(sendableDirRoot, f.Dir())
		if err != nil {
			return err
		}
	}
	fInfo, err := anypb.New(&kanbanpb.FileInfo{
		Size:     f.Size(),
		ChunkCnt: chunkCnt,
		Hash:     nil,
		Name:     f.Name(),
		RelDir:   relDir,
	})
	if err != nil {
		return err
	}

	err = sendToServer(stream, &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_SendingFile_EOF,
		Context: fInfo,
	})

	if err != nil {
		return err
	}
	return nil
}

// send eos to other devices
func sendEOS(stream kanbanpb.SendAnything_SendToOtherDevicesClient, fileCnt int) (err error) {
	eos, err := anypb.New(&kanbanpb.StreamInfo{
		FileCount: int32(fileCnt),
	})
	if err != nil {
		defer func(e error) {
			if err == nil {
				err = e
			}
		}(err)
	}
	req := &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_EOS,
		Context: eos,
	}
	if err := stream.Send(req); err != nil {
		return err
	}
	return nil
}

func LOG(any ...interface{}) {
	log.Println(any...)
	return
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
