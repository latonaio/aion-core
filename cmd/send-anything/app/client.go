package app

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"time"

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

func sendToOtherDeviceClient(ctx context.Context, m *kanbanpb.SendKanban, port int) error {
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

	// send kanban and files
	if err := startSending(stream, m); err != nil {
		return errors.Wrap(err, fmt.Sprintf("destination　address: %s", deviceAddr))
	}

	return nil
}

func startSending(stream kanbanpb.SendAnything_SendToOtherDevicesClient, m *kanbanpb.SendKanban) error {
	files, err := baggage.CreateFilesInfo(m.AfterKanban.FileList)
	if err != nil {
		return err
	}
	// 全て送るか、全て送らないかにしている。後で、選択できるようにしたい。
	allOK := files.IsAllExist()

	if !allOK {
		invalidFilePaths := make([]string, 0, len(files))
		for i, f := range files {
			if !f.IsExist() {
				invalidFilePaths = append(invalidFilePaths, m.AfterKanban.FileList[i])
			}
		}
		return errors.Errorf("cannot open file(s) %v", invalidFilePaths)
	}

	errLog := make([]error, 0, 3)

	if err := sendKanban(stream, m); err != nil {
		errLog = append(errLog, errors.Wrap(err, "[SendToOtherDevice client] sending kanban is failed"))
	}
	if err := sendFileList(stream, m, files); err != nil {
		errLog = append(errLog, errors.Wrap(err, "[SendToOtherDevice client] sending files are failed"))
	}
	if err := sendEos(stream); err != nil {
		errLog = append(errLog, errors.Wrap(err, "[SendToOtherDevice client] sending eos is failed"))
	}
	if len(errLog) > 0 {
		err = errors.Errorf("sending error: %v", errLog)
	}
	return err
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

// send data by file list to other devices
func sendFileList(stream kanbanpb.SendAnything_SendToOtherDevicesClient, m *kanbanpb.SendKanban, files baggage.FilesInfo) error {
	errStr := make([]string, 0, len(files))
	for _, f := range files {
		if f != nil {
			if err := sendFile(stream, f); err != nil {
				errStr = append(errStr, err.Error())
			}
		}
	}
	if len(errStr) != 0 {
		return fmt.Errorf("cant send files\n%s", strings.Join(errStr, "\n"))
	}
	return nil
}

func sendFileInfo(stream kanbanpb.SendAnything_SendToOtherDevicesClient, f *baggage.FileInfo) error {
	req := &kanbanpb.SendContext{}
	req.Code = kanbanpb.UploadRequestCode_SendingFile_Info
	hash, err := f.Hash()
	if err != nil {
		return err
	}

	anyInfo, err := anypb.New(
		&kanbanpb.FileInfo{
			Size:     int32(f.Size()),
			Hash:     hash,
			Name:     f.Name(),
			ChunkCnt: int32(f.Size() / chunkSize),
		},
	)
	if err != nil {
		return err
	}

	req.Context = anyInfo
	if err := sendToServer(stream, req); err != nil {
		return sendToServerNotifyFailure(stream)
	}
	return nil
}

// send data to other devices
func sendFile(stream kanbanpb.SendAnything_SendToOtherDevicesClient, file *baggage.FileInfo) error {
	if file == nil || !file.IsExist() {
		return errors.Errorf("skip sending a file")
	}
	log.Printf("[SendToOtherDevices client] start to send file: %s", file.Name())

	err := sendFileInfo(stream, file)
	if err != nil {
		return err
	}
	for refNum := int32(0); true; refNum++ {
		var req *kanbanpb.SendContext
		chunk := make([]byte, chunkSize)
		count, err := file.GetChunk(chunk)
		if err != nil {
			if err != io.EOF {
				return sendToServerNotifyFailure(stream)
			}
			anyChunk, err := anypb.New(&kanbanpb.Chunk{
				Context: chunk[:count],
				Name:    file.Path(),
				RefNum:  refNum,
			})
			if err != nil {
				return sendToServerNotifyFailure(stream)
			}
			req = &kanbanpb.SendContext{
				Code:    kanbanpb.UploadRequestCode_SendingFile_EOF,
				Context: anyChunk,
			}
			if err := sendToServer(stream, req); err != nil {
				return sendToServerNotifyFailure(stream)
			}
			break
		}
		anyChunk, err := anypb.New(&kanbanpb.Chunk{
			Context: chunk[:count],
			Name:    file.Path(),
			RefNum:  refNum,
		})
		if err != nil {
			return sendToServerNotifyFailure(stream)
		}
		req = &kanbanpb.SendContext{
			Code:    kanbanpb.UploadRequestCode_SendingFile_CONT,
			Context: anyChunk,
		}
		if err := sendToServer(stream, req); err != nil {
			return sendToServerNotifyFailure(stream)
		}
	}
	log.Printf("[SendToOtherDevices client] finish to send file: %s", file.Name())
	return nil
}

// send eos to other devices
func sendEos(stream kanbanpb.SendAnything_SendToOtherDevicesClient) error {
	req := &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_EOS,
		Context: nil,
	}
	if err := stream.Send(req); err != nil {
		return err
	}
	return nil
}

func LOG(any interface{}) {
	_, f, l, _ := runtime.Caller(1)
	log.Printf("DEBUG from %s:%d", f, l)
	log.Printf("%v", any)
	log.Printf("")
}
