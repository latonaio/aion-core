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
	nextRefNum := int32(0)

	// output file var
	var recvFileBuf []byte
	numOfOutputFile := 0
	hash := []byte{}

	for {
		// wait stream by client
		in, err := stream.Recv()
		if err != nil {
			return fmt.Sprintf("cant receive message, %v", err), kanbanpb.UploadStatusCode_Failed
		}
		fmt.Printf("%s", in.Code)
		switch in.Code {
		// receive kanban
		case kanbanpb.UploadRequestCode_SendingKanban:
			// receive file
			if err := sendingKanban(in, kContainer, &dirPath, &nextRefNum); err != nil {
				return err.Error(), kanbanpb.UploadStatusCode_Failed
			}
		case kanbanpb.UploadRequestCode_SendingFile_Info:
			// receive file
			if err := sendingKanbanInfo(in, &hash); err != nil {
				return err.Error(), kanbanpb.UploadStatusCode_Failed
			}
		case kanbanpb.UploadRequestCode_SendingFile_CONT:
			if err := sendingKanbanCont(in.Context, &recvFileBuf, nextRefNum); err != nil {
				return err.Error(), kanbanpb.UploadStatusCode_Failed
			}
			nextRefNum++
		case kanbanpb.UploadRequestCode_SendingFile_EOF:
			// receive eos
			if err := sendingFileEOF(in, recvFileBuf, dirPath, &numOfOutputFile, hash, &nextRefNum); err != nil {
				return err.Error(), kanbanpb.UploadStatusCode_Failed
			}
			recvFileBuf = make([]byte, 0, 0)
		case kanbanpb.UploadRequestCode_EOS:
			// receive eof from stream
			if numOfOutputFile != len(kContainer.AfterKanban.FileList) {
				return "enough files haven't received yet", kanbanpb.UploadStatusCode_Failed
			}
			srv.sendToServiceBrokerCh <- kContainer
			return "all files are received", kanbanpb.UploadStatusCode_OK
		default:
			if numOfOutputFile != len(kContainer.AfterKanban.FileList) {
				return "enough files haven't received yet", kanbanpb.UploadStatusCode_Failed
			}
			srv.sendToServiceBrokerCh <- kContainer
			return "all files are received", kanbanpb.UploadStatusCode_OK
		}
	}
}
func sendingKanban(in *kanbanpb.SendContext, kContainer *kanbanpb.SendKanban, dirPath *string, lastRefNum *int32) error {
	if err := ptypes.UnmarshalAny(in.Context, kContainer); err != nil {
		return fmt.Errorf("cant unmarshal any, %v", err)
	}
	*dirPath = common.GetMsDataPath(*dirPath, kContainer.NextService, int(kContainer.NextNumber))
	if f, err := os.Stat(*dirPath); os.IsNotExist(err) || !f.IsDir() {
		if err := os.Mkdir(*dirPath, 0775); err != nil {
			log.Printf("[SendToOtherDevices server] cant create directory, %v", err)
		}
	}
	log.Printf("[SendToOtherDevices server] receive kanban (nextService: %s)", kContainer.NextService)
	return nil
}
func sendingKanbanInfo(in *kanbanpb.SendContext, hash *[]byte) error {
	content := &kanbanpb.FileInfo{}
	if err := ptypes.UnmarshalAny(in.Context, content); err != nil {
		return fmt.Errorf("cant unmarshal any, %v", err)
	}

	// ちゃんととって来れているか確認
	LOG(content)
	*hash = content.Hash
	return nil
}
func sendingKanbanCont(dataContent *anypb.Any, recvFileBuf *[]byte, wantRefNum int32) error {
	chunk := &kanbanpb.Chunk{}
	if err := ptypes.UnmarshalAny(dataContent, chunk); err != nil {
		return fmt.Errorf("cant unmarshal any, %v", err)
	}
	if chunk.RefNum != wantRefNum {
		return fmt.Errorf("sending kanban order is not correct. want:%d got:%d", wantRefNum, chunk.RefNum)
	}
	wantRefNum++
	*recvFileBuf = append(*recvFileBuf, chunk.Context...)
	return nil
}

func sendingFileEOF(in *kanbanpb.SendContext, recvFileBuf []byte, dirPath string, numOfOutputFile *int, hash []byte, nextRefNum *int32) error {
	chunk := &kanbanpb.Chunk{}
	if err := ptypes.UnmarshalAny(in.Context, chunk); err != nil {
		return fmt.Errorf("cant unmarshal any, %v", err)
	}
	fileName := filepath.Base(chunk.Name)
	filePath := path.Join(dirPath, fileName)

	if f, err := os.Stat(dirPath); os.IsNotExist(err) || !f.IsDir() {
		if err := os.MkdirAll(dirPath, 0775); err != nil {
			log.Printf("[SendToOtherDevices server] cant create directory, %v", err)
		}
	}
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("cant open output path: %v", err)
	}
	if _, err := f.Write(recvFileBuf); err != nil {
		f.Close()
		return fmt.Errorf("cant write output file: %v", err)
	}

	*nextRefNum = 0
	f.Seek(0, 0)

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	if string(hash) != string(h.Sum(nil)) {
		return fmt.Errorf("Received file is broken")
	}

	log.Printf("[SendToOtherDevices server] success to write file (%s)", filePath)
	*numOfOutputFile += 1
	f.Close()
	return nil
}
