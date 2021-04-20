package app

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"syscall"

	"bitbucket.org/latonaio/aion-core/pkg/common"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
)

// chunk size (128kB)
const (
	chunkSize = 128 * 1024
)

type Server struct {
	env                   *Env
	sendToServiceBrokerCh chan *kanbanpb.SendKanban
}

// create struct of server and call createServerConnection
func NewServer(env *Env) error {
	server := &Server{
		env:                   env,
		sendToServiceBrokerCh: make(chan *kanbanpb.SendKanban, 1),
	}
	return server.createServerConnection()
}

// start send anything server
func (srv *Server) createServerConnection() error {
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(srv.env.GetServerPort()))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	kanbanpb.RegisterSendAnythingServer(grpcServer, srv)
	log.Printf("Start send anything server:%d", srv.env.GetServerPort())

	errCh := make(chan error)
	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	select {
	case err := <-errCh:
		return err
	case <-quit:
		grpcServer.GracefulStop()
		return nil
	}
}

// callback that connect with service broker
func (srv *Server) ServiceBrokerConn(stream kanbanpb.SendAnything_ServiceBrokerConnServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	log.Printf("connect from servicebroker")

	go func() {
		for {
			// wait stream by client
			in, err := stream.Recv()
			if err != nil {
				log.Printf("receive stream is closed: %v", err)
				return
			}
			log.Printf("[ServiceBrokerConn] receive from service broker")
			if err := sendToOtherDeviceClient(ctx, in, srv.env.ClientPort); err != nil {
				log.Printf("%v", err)
			}
		}
	}()

	// start loop that send request to service broker
	for {
		select {
		case <-ctx.Done():
			return nil
		case k, ok := <-srv.sendToServiceBrokerCh:
			if !ok {
				return nil
			}
			if err := stream.Send(k); err != nil {
				log.Printf("[ServiceBrokerConn] grpc send error: %v", err)
			} else {
				log.Printf("[ServiceBrokerConn] send to servicebroker (from device: %s, next service: %s)",
					k.DeviceName, k.NextService)
			}
		}
	}
}

// callback that server receive data from other client
func (srv *Server) SendToOtherDevices(stream kanbanpb.SendAnything_SendToOtherDevicesServer) error {
	log.Printf("[SendToOtherDevices server] create connection")
	message, code := srv.messageParserInSendAnythingServer(stream, srv.env.GetDataDir())
	LOG(message)
	resp := &kanbanpb.UploadStatus{
		Message:    "[SendToOtherDevices server] " + message,
		StatusCode: code,
	}
	if err := stream.SendAndClose(resp); err != nil {
		return errors.Wrap(err, "[SendToOtherDevices server] sending result message is failed: %v")
	}
	return nil
}

// parsing message that received from send anything client in remote devices
func (srv *Server) messageParserInSendAnythingServer(stream kanbanpb.SendAnything_SendToOtherDevicesServer, dirPath string) (string, kanbanpb.UploadStatusCode) {

	// received message var
	kContainer := &kanbanpb.SendKanban{}

	// output file var
	fileInfo := &kanbanpb.FileInfo{}
	fileBuilder := new(receivingFileBuilder)
	receivedFileCnt := 0
	outputFiles := []string{}

	for {
		// wait stream by client
		in, err := stream.Recv()
		if err != nil {
			return fmt.Sprintf("cant receive message, %v", err), kanbanpb.UploadStatusCode_Failed
		}
		switch in.Code {
		// receive kanban
		case kanbanpb.UploadRequestCode_SendingKanban:
			kContainer, err = receiveSendingKanban(in, &dirPath)
			if err != nil {
				return err.Error(), kanbanpb.UploadStatusCode_Failed
			}
			outputFiles = make([]string, 0, len(kContainer.AfterKanban.FileList))
		case kanbanpb.UploadRequestCode_SendingFile_Info:
			fileInfo, fileBuilder, err = receiveFileInfo(in)
			if err != nil {
				return err.Error(), kanbanpb.UploadStatusCode_Failed
			}
		case kanbanpb.UploadRequestCode_SendingFile_CONT:
			if err = receiveFileCont(in.Context, fileBuilder); err != nil {
				return err.Error(), kanbanpb.UploadStatusCode_Failed
			}
		case kanbanpb.UploadRequestCode_SendingFile_EOF:
			if err = receiveEOF(in, fileInfo, fileBuilder, dirPath, &outputFiles, &receivedFileCnt); err != nil {
				return err.Error(), kanbanpb.UploadStatusCode_Failed
			}
			LOG(outputFiles)
		case kanbanpb.UploadRequestCode_EOS:
			if err = receiveEOS(in, kContainer, &receivedFileCnt); err != nil {
				return "enough files haven't received yet", kanbanpb.UploadStatusCode_Failed
			}
			kContainer.AfterKanban.FileList = outputFiles
			srv.sendToServiceBrokerCh <- kContainer
			return "all files are received", kanbanpb.UploadStatusCode_OK
		default:
			if receivedFileCnt != len(kContainer.AfterKanban.FileList) {
				return "enough files haven't received yet", kanbanpb.UploadStatusCode_Failed
			}
			srv.sendToServiceBrokerCh <- kContainer
			return "all files are received", kanbanpb.UploadStatusCode_OK
		}
	}
}

func receiveSendingKanban(in *kanbanpb.SendContext, dirPath *string) (*kanbanpb.SendKanban, error) {
	kContainer := &kanbanpb.SendKanban{}

	if err := ptypes.UnmarshalAny(in.Context, kContainer); err != nil {
		return nil, fmt.Errorf("cant unmarshal any, %v", err)
	}
	*dirPath = common.GetMsDataPath(*dirPath, kContainer.NextService, int(kContainer.NextNumber))
	if f, err := os.Stat(*dirPath); os.IsNotExist(err) || !f.IsDir() {
		if err := os.Mkdir(*dirPath, 0775); err != nil {
			log.Printf("[SendToOtherDevices server] cant create directory, %v", err)
		}
	}
	log.Printf("[SendToOtherDevices server] receive kanban (nextService: %s)", kContainer.NextService)
	return kContainer, nil
}
func receiveFileInfo(in *kanbanpb.SendContext) (*kanbanpb.FileInfo, *receivingFileBuilder, error) {
	content := &kanbanpb.FileInfo{}
	if err := ptypes.UnmarshalAny(in.Context, content); err != nil {
		return nil, nil, fmt.Errorf("cant unmarshal any, %v", err)
	}

	// ちゃんととって来れているか確認
	LOG(content)
	fileBuilder := newReceivingFileBuilder(content)

	return content, fileBuilder, nil
}
func receiveFileCont(dataContent *anypb.Any, file *receivingFileBuilder) error {
	chunk := &kanbanpb.Chunk{}
	if err := ptypes.UnmarshalAny(dataContent, chunk); err != nil {
		return fmt.Errorf("cant unmarshal any, %v", err)
	}
	// LOG(fmt.Sprintf("chunkLen:%d , Want refNum:%d", len(chunk.Context), file.expectRefefNum))
	if chunk.RefNum != file.expectRefefNum {
		return fmt.Errorf("sending kanban order is not correct. want:%d got:%d", file.expectRefefNum, chunk.RefNum)
	}
	file.expectRefefNum++
	file.stack(&chunk.Context)
	return nil
}

func receiveEOF(in *kanbanpb.SendContext, fInfo *kanbanpb.FileInfo, file *receivingFileBuilder, dirPath string, outputFiles *[]string, numOfOutputFile *int) error {
	file.expectRefefNum = 0
	fInfoEOF := &kanbanpb.FileInfo{}
	if err := ptypes.UnmarshalAny(in.Context, fInfoEOF); err != nil {
		return fmt.Errorf("cant unmarshal any, %v", err)
	}
	fileName := filepath.Base(fInfoEOF.Name)
	outputDir := path.Join(dirPath, fInfoEOF.RelDir)
	filePath := path.Join(outputDir, fileName)

	if f, err := os.Stat(outputDir); os.IsNotExist(err) || !f.IsDir() {
		if err := os.MkdirAll(outputDir, 0775); err != nil {
			log.Printf("[SendToOtherDevices server] cant create directory, %v", err)
		}
	}
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("cant open output path: %v", err)
	}
	defer f.Close()
	if _, err := f.Write(file.fileContent); err != nil {
		return fmt.Errorf("cant write output file: %v", err)
	}
	*outputFiles = append(*outputFiles, filePath)

	if string(fInfo.Hash) != getMD5(f) {
		os.Remove(f.Name())
		return fmt.Errorf("Received file [%v] is broken", f.Name())
	}

	log.Printf("[SendToOtherDevices server] success to write file (%s)", filePath)
	*numOfOutputFile += 1
	return nil
}

func getMD5(f *os.File) string {
	f.Seek(0, 0)
	defer f.Seek(0, 0)

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return string(h.Sum(nil))
}

func receiveEOS(in *kanbanpb.SendContext, kanban *kanbanpb.SendKanban, numOfOutputFile *int) error {
	eos := &kanbanpb.StreamInfo{}
	if err := ptypes.UnmarshalAny(in.Context, eos); err != nil {
		return fmt.Errorf("cant unmarshal any, %v", err)
	}

	if eos.FileCount != int32(*numOfOutputFile) {
		return errors.Errorf("sent files count and received files count is NOT equal. got: %d, want: %d", *numOfOutputFile, eos.FileCount)
	}

	*numOfOutputFile = 0
	return nil
}

type receivingFileBuilder struct {
	name           string
	size           int64
	fileContent    []byte
	expectRefefNum int32
}

func newReceivingFileBuilder(fInfo *kanbanpb.FileInfo) *receivingFileBuilder {
	content := make([]byte, 0, fInfo.Size)
	fb := &receivingFileBuilder{
		name:           fInfo.Name,
		size:           fInfo.Size,
		fileContent:    content,
		expectRefefNum: 0,
	}
	return fb
}

func (fb *receivingFileBuilder) stack(content *[]byte) {
	fb.fileContent = append(fb.fileContent, *content...)
}
