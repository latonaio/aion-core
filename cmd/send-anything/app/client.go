package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

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

	eachFileStatus, isAllOK := filesExists(m.AfterKanban.FileList)

	if !isAllOK {
		invalidFilePaths := make([]string, 0, len(eachFileStatus))
		for i, ok := range eachFileStatus {
			if !ok {
				invalidFilePaths = append(invalidFilePaths, m.AfterKanban.FileList[i])
			}
		}
		return errors.Errorf("cannot open file(s) %v", invalidFilePaths)
	}

	// send kanban and files
	if err := sendKanban(stream, m); err != nil {
		return errors.Wrap(err, fmt.Sprintf("[SendToOtherDevice client] sending kanban is failed: %s", deviceAddr))
	}
	if err := sendFileList(stream, m); err != nil {
		return errors.Wrap(err, fmt.Sprintf("[SendToOtherDevice client] sending files are failed: %s", deviceAddr))
	}
	if err := sendEos(stream); err != nil {
		return errors.Wrap(err, fmt.Sprintf("[SendToOtherDevice client] sending eos is failed: %s", deviceAddr))
	}

	return nil
}

func filesExists(filePaths []string) ([]bool, bool) {
	existStatus := make([]bool, 0, len(filePaths))
	allOK := true

	for _, path := range filePaths {
		thisFile := fileExists(path)
		existStatus = append(existStatus, thisFile)

		if allOK && !thisFile {
			allOK = false
		}
	}
	return existStatus, allOK
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

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// send data to other devices
func sendFile(stream kanbanpb.SendAnything_SendToOtherDevicesClient, fileName string, filePath string) error {
	log.Printf("[SendToOtherDevices client] start to send file: %s, %s", fileName, filePath)
	if err := checkFileStatus(filePath); err != nil {
		return err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cant open file: %s", filePath))
	}
	defer f.Close()

	for cnt := int32(0); true; cnt++ {
		var req *kanbanpb.SendContext
		chunk := make([]byte, chunkSize)
		count, err := f.Read(chunk)
		if err != nil {
			if err != io.EOF {
				return sendToServerNotifyFailure(stream)
			}
			req = &kanbanpb.SendContext{
				Code:    kanbanpb.UploadRequestCode_SendingFile_EOF,
				Context: nil,
			}
			if err := sendToServer(stream, req); err != nil {
				return sendToServerNotifyFailure(stream)
			}
			break
		}
		anyChunk, err := anypb.New(&kanbanpb.Chunk{
			Context: chunk[:count],
			Name:    fileName,
			RefNum:  cnt,
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
	log.Printf("[SendToOtherDevices client] finish to send file: %s, %s", fileName, filePath)
	return nil
}

// send data by file list to other devices
func sendFileList(stream kanbanpb.SendAnything_SendToOtherDevicesClient, m *kanbanpb.SendKanban) error {
	var errStr []string
	for _, fileName := range m.AfterKanban.FileList {
		fileExists := fileExists(fileName)
		if fileExists {
			filePath := fileName
			if err := sendFile(stream, fileName, filePath); err != nil {
				errStr = append(errStr, err.Error())
			}
		} else {
			filePath := path.Join(m.AfterKanban.DataPath, fileName)
			if err := sendFile(stream, fileName, filePath); err != nil {
				errStr = append(errStr, err.Error())
			}
		}
	}
	if len(errStr) != 0 {
		return fmt.Errorf("cant send files\n%s", strings.Join(errStr, "\n"))
	}
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
