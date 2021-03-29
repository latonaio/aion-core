package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"bitbucket.org/latonaio/aion-core/pkg/common"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
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
			if err := sendToOtherDeviceClient(ctx, in, srv.env.ServerPort); err != nil {
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

func checkFileStatus(filePath string) error {
	var fileSize int64 = 0
	for i := 0; i < 10; i++ {
		f, err := os.Open(filePath)
		defer f.Close()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("cant open file: %s", filePath))
		}
		stat, err := f.Stat()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("cant get file stat: %s", filePath))
		}
		if i == 0 {
			fileSize = stat.Size()
		} else {
			prevFileSize := fileSize
			fileSize = stat.Size()
			log.Printf("[SendToOtherDevices client] check file size: %v", fileSize)
			if prevFileSize == fileSize && fileSize != 0 {
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

// callback that server receive data from other client
func (srv *Server) SendToOtherDevices(stream kanbanpb.SendAnything_SendToOtherDevicesServer) error {
	log.Printf("[SendToOtherDevices server] create connection")
	message, code := srv.messageParserInSendAnythingServer(stream, srv.env.GetDataDir())
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
	chunk := &kanbanpb.Chunk{}

	// output file var
	var recvFileBuf []byte
	numOfOutputFile := 0

	for {
		// wait stream by client
		in, err := stream.Recv()
		if err != nil {
			return fmt.Sprintf("cant receive message"), kanbanpb.UploadStatusCode_Failed
		}
		fmt.Printf("%s", in.Code)
		switch in.Code {
		// receive kanban
		case kanbanpb.UploadRequestCode_SendingKanban:
			if err := ptypes.UnmarshalAny(in.Context, kContainer); err != nil {
				return fmt.Sprintf("cant unmarshal any, %v", err), kanbanpb.UploadStatusCode_Failed
			}
			dirPath = common.GetMsDataPath(dirPath, kContainer.NextService, int(kContainer.NextNumber))
			if f, err := os.Stat(dirPath); os.IsNotExist(err) || !f.IsDir() {
				if err := os.Mkdir(dirPath, 0775); err != nil {
					log.Printf("[SendToOtherDevices server] cant create directory, %v", err)
				}
			}
			log.Printf("[SendToOtherDevices server] receive kanban (nextService: %s)", kContainer.NextService)

			// receive file
		case kanbanpb.UploadRequestCode_SendingFile_CONT:
			if err := ptypes.UnmarshalAny(in.Context, chunk); err != nil {
				return fmt.Sprintf("cant unmarshal any, %v", err), kanbanpb.UploadStatusCode_Failed
			}
			recvFileBuf = append(recvFileBuf, chunk.Context...)
		case kanbanpb.UploadRequestCode_SendingFile_EOF:
			fileName := filepath.Base(chunk.Name)
			filePath := path.Join(dirPath, fileName)
			if f, err := os.Stat(dirPath); os.IsNotExist(err) || !f.IsDir() {
				if err := os.MkdirAll(dirPath, 0775); err != nil {
					log.Printf("[SendToOtherDevices server] cant create directory, %v", err)
				}
			}
			f, err := os.Create(filePath)
			if err != nil {
				return fmt.Sprintf("cant open output path: %v", err),
					kanbanpb.UploadStatusCode_Failed
			}
			if _, err := f.Write(recvFileBuf); err != nil {
				f.Close()
				return fmt.Sprintf("cant write output file: %v", err),
					kanbanpb.UploadStatusCode_Failed
			}
			log.Printf("[SendToOtherDevices server] success to write file (%s)", filePath)
			fileName = ""
			recvFileBuf = []byte{}
			numOfOutputFile += 1
			f.Close()
		// receive eos
		case kanbanpb.UploadRequestCode_EOS:
			if numOfOutputFile != len(kContainer.AfterKanban.FileList) {
				return fmt.Sprintf("enough files haven't received yet"), kanbanpb.UploadStatusCode_Failed
			}
			srv.sendToServiceBrokerCh <- kContainer
			return fmt.Sprintf("all files are received"), kanbanpb.UploadStatusCode_OK
		// receive eof from stream
		default:
			if numOfOutputFile != len(kContainer.AfterKanban.FileList) {
				return fmt.Sprintf("enough files haven't received yet"), kanbanpb.UploadStatusCode_Failed
			}
			srv.sendToServiceBrokerCh <- kContainer
			return fmt.Sprintf("all files are received"), kanbanpb.UploadStatusCode_OK
		}
	}
}
