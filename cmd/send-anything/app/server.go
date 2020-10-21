package app

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/latonaio/aion-core/pkg/common"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"github.com/avast/retry-go"
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
	return grpcServer.Serve(listen)
}

// callback that connect with service broker
func (srv *Server) ServiceBrokerConn(stream kanbanpb.SendAnything_ServiceBrokerConnServer) error {
	ctx := stream.Context()
	log.Printf("connect from servicebroker")

	// create receive channel
	recvCh := make(chan *kanbanpb.SendKanban, 1)
	go func() {
		for {
			// wait stream by client
			in, err := stream.Recv()
			if err != nil {
				log.Printf("receive stream is closed: %v", err)
				close(recvCh)
				return
			}
			log.Printf("[ServiceBrokerConn] receive from service broker")
			recvCh <- in
		}
	}()

	// start loop that send request to service broker
	go func() {
		for k := range srv.sendToServiceBrokerCh {
			if err := stream.Send(k); err != nil {
				log.Printf("[ServiceBrokerConn] grpc send error: %v", err)
			} else {
				log.Printf("[ServiceBrokerConn] send to servicebroker (from device: %s, next service: %s)",
					k.DeviceName, k.NextService)
			}
		}
	}()

	// loop in receive ch and ctx Done
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case m, ok := <-recvCh:
			if !ok {
				break loop
			}
			// connect to send anything in other devices amd send data
			if err := srv.sendToOtherDeviceClient(ctx, m); err != nil {
				log.Printf("%v", err)
			}
		}
	}
	return nil
}

// connect to send anything in other devices amd send data
func (srv *Server) sendToOtherDeviceClient(ctx context.Context, m *kanbanpb.SendKanban) error {
	// connect to remote device
	deviceAddr := m.DeviceAddr + ":" + strconv.Itoa(srv.env.GetClientPort())
	conn, err := grpc.DialContext(ctx, deviceAddr, grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("cannot connect to remote device: %s", deviceAddr))
		log.Print(err)
		return err
	}
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

// send kanban to other devices
func sendKanban(stream kanbanpb.SendAnything_SendToOtherDevicesClient, m *kanbanpb.SendKanban) error {
	log.Printf("[SendToOtherDevices client] start to send kanban: %s", m.NextService)
	cont, err := ptypes.MarshalAny(m)
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

// send data to other devices
func sendFile(stream kanbanpb.SendAnything_SendToOtherDevicesClient, fileName string, filePath string) error {
	log.Printf("[SendToOtherDevices client] start to send file: %s, %s", fileName, filePath)
	if err := checkFileStatus(filePath); err != nil {
		return err
	}
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cant open file: %s", filePath))
	}

	req := &kanbanpb.SendContext{
		Code:    kanbanpb.UploadRequestCode_SendingFile_CONT,
		Context: nil,
	}
	chunk := &kanbanpb.Chunk{
		Context: make([]byte, chunkSize),
		Name:    fileName,
	}

	stat, err := f.Stat()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cant get file stat: %s", filePath))
	}
	fileBuf := make([]byte, stat.Size())
	if _, err := f.Read(fileBuf); err != nil {
		return errors.Wrap(err, fmt.Sprintf("cant read file: %s", filePath))
	}
	for currentByte := 0; currentByte < len(fileBuf); currentByte += chunkSize {
		if currentByte+chunkSize > len(fileBuf) {
			req.Code = kanbanpb.UploadRequestCode_SendingFile_EOF
			chunk.Context = fileBuf[currentByte:]
		} else {
			chunk.Context = fileBuf[currentByte : currentByte+chunkSize]
		}
		anyChunk, err := ptypes.MarshalAny(chunk)
		if err != nil {
			return errors.Wrap(err, "cant marshal to any")
		}
		req.Context = anyChunk
		const connRetryCount = 10
		if err := retry.Do(
			func() error {
				if err := stream.Send(req); err != nil {
					log.Printf("[SendToOtherDevices client] failed to send chunk (error: %v)", err)
					return err
				}
				return nil
			},
			retry.DelayType(func(n uint, config *retry.Config) time.Duration {
				log.Printf("[SendToOtherDevices client] retry to send because failed to send buffer ")
				return time.Second
			}),
			retry.Attempts(connRetryCount),
		); err != nil {
			return err
		}
	}
	log.Printf("[SendToOtherDevices client] finish to send file: %s, %s", fileName, filePath)
	return nil
}

// callback that server receive data from other client
func (srv *Server) SendToOtherDevices(stream kanbanpb.SendAnything_SendToOtherDevicesServer) error {
	log.Printf("[SendToOtherDevices server] create connection")
	ctx := stream.Context()
	message, code := srv.messageParserInSendAnythingServer(ctx, stream, srv.env.GetDataDir())
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
func (srv *Server) messageParserInSendAnythingServer(ctx context.Context,
	stream kanbanpb.SendAnything_SendToOtherDevicesServer, dirPath string) (string, kanbanpb.UploadStatusCode) {

	// received message var
	kContainer := &kanbanpb.SendKanban{}
	chunk := &kanbanpb.Chunk{}

	// output file var
	var recvFileBuf []byte
	var fileName string

	numOfOutputFile := 0

	recvCh := make(chan *kanbanpb.SendContext, 1)
	go func() {
		for {
			// wait stream by client
			in, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Printf("[SendToOtherDevices server] received err: %v", err)
				}
				recvCh <- &kanbanpb.SendContext{Code: -1}
				return
			}
			fmt.Printf("%s", in.Code)
			recvCh <- in
		}
	}()

	for {
		// wait stream by client
		select {
		case <-ctx.Done():
			return fmt.Sprintf("cant receive message"), kanbanpb.UploadStatusCode_Failed
		case in := <-recvCh:
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
			case kanbanpb.UploadRequestCode_SendingFile_CONT, kanbanpb.UploadRequestCode_SendingFile_EOF:
				if err := ptypes.UnmarshalAny(in.Context, chunk); err != nil {
					return fmt.Sprintf("cant unmarshal any, %v", err), kanbanpb.UploadStatusCode_Failed
				}
				fileName = ""
				if fileName == "" {
					fileName = chunk.Name
				} else if fileName != chunk.Name {
					return fmt.Sprintf("receive other file before eos received (before: %s, after: %s)",
						fileName, chunk.Name), kanbanpb.UploadStatusCode_Failed
				}
				recvFileBuf = append(recvFileBuf, chunk.Context...)
				if in.Code == kanbanpb.UploadRequestCode_SendingFile_EOF {
					fileName := filepath.Base(fileName)
					filePath := path.Join(dirPath, fileName)
					fileDirPath := path.Dir(filePath)
					if f, err := os.Stat(fileDirPath); os.IsNotExist(err) || !f.IsDir() {
						if err := os.MkdirAll(fileDirPath, 0775); err != nil {
							log.Printf("[SendToOtherDevices server] cant create directory, %v", err)
						}
					}
					f, err := os.Create(filePath)
					if err != nil {
						return fmt.Sprintf("cant open output path: %v", err),
							kanbanpb.UploadStatusCode_Failed
					}
					if _, err := f.Write(recvFileBuf); err != nil {
						return fmt.Sprintf("cant write output file: %v", err),
							kanbanpb.UploadStatusCode_Failed
					}
					log.Printf("[SendToOtherDevices server] success to write file (%s)", filePath)
					fileName = ""
					recvFileBuf = []byte{}
					numOfOutputFile += 1
				}
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
}
