package app

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"bitbucket.org/latonaio/aion-core/pkg/common"
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/xerrors"
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
		return xerrors.Errorf("failed to listen: %w", err)
	}
	grpcServer := grpc.NewServer()
	kanbanpb.RegisterSendAnythingServer(grpcServer, srv)
	log.Printf("Start send anything server:%d", srv.env.GetServerPort())

	errCh := make(chan error)
	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			errCh <- xerrors.Errorf("serve error: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	select {
	case err := <-errCh:
		return xerrors.Errorf("receive error channel: %w", err)
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
				log.Printf("receive stream is closed: %+v", err)
				return
			}
			log.Printf("[ServiceBrokerConn] receive from service broker")
			if err := sendToOtherDeviceClient(ctx, in, srv.env.ClientPort); err != nil {
				log.Printf("%+v", err)
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
				log.Printf("[ServiceBrokerConn] grpc send error: %+v", err)
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
		return xerrors.Errorf("[SendToOtherDevices server] sending result message is failed: %w", err)
	}
	return nil
}

// parsing message that received from send anything client in remote devices
func (srv *Server) messageParserInSendAnythingServer(stream kanbanpb.SendAnything_SendToOtherDevicesServer, dirPath string) (string, kanbanpb.UploadStatusCode) {
	// received message var
	kContainer := &kanbanpb.SendKanban{}
	// 最後に出力するkanbanの作成に必要
	in, err := stream.Recv()
	if err != nil {
		return fmt.Sprintf("cant receive message, %+v", err), kanbanpb.UploadStatusCode_Failed
	}
	if in.Code != kanbanpb.UploadRequestCode_SendingKanban {
		return xerrors.Errorf("got unknown code %v", in.Code).Error(), kanbanpb.UploadStatusCode_Failed
	}

	kContainer, err = receiveSendingKanban(in, &dirPath)
	if err != nil {
		return err.Error(), kanbanpb.UploadStatusCode_Failed
	}

	// kanban送信のみのリクエストをここで終わらせるべきか？
	defer func() {
		srv.sendToServiceBrokerCh <- kContainer
	}()
	if err = downloadAnything(stream, kContainer, dirPath); err != nil {
		LOG(fmt.Sprintf("%+v\n", err))
		return err.Error(), kanbanpb.UploadStatusCode_Failed
	}
	return "ok", kanbanpb.UploadStatusCode_OK
}

func downloadAnything(stream kanbanpb.SendAnything_SendToOtherDevicesServer, kContainer *kanbanpb.SendKanban, outputDirPath string) error {
	itemInfo := newOutputItemInfo()

loop:
	for {
		in, err := stream.Recv()
		if err != nil {
			return xerrors.Errorf("failed receiving kanban: %w", err)
		}
		switch in.Code {
		case kanbanpb.UploadRequestCode_SendingFile_Info:
			f, err := receiveFileInfo(stream, in, outputDirPath)
			if err != nil {
				return xerrors.Errorf("failed receiving file info: %w", err)
			}
			itemInfo.files = append(itemInfo.files, f)
		case kanbanpb.UploadRequestCode_SendingDirInfo:
			d, err := receiveDirInfo(stream, in, outputDirPath)
			if err != nil {
				return xerrors.Errorf("failed receiving directory info: %w", err)
			}
			itemInfo.dirs = append(itemInfo.dirs, d)
		case kanbanpb.UploadRequestCode_EOS:
			if err = receiveEOS(in, itemInfo); err != nil {
				return xerrors.Errorf("enough files haven't received yet: %w", err)
			}
			kContainer.AfterKanban.FileList = itemInfo.list()
			break loop
		default:
			if len(itemInfo.list()) != len(kContainer.AfterKanban.FileList) {
				return xerrors.Errorf("enough files haven't received yet: get: %d, want: %d", len(itemInfo.list()), len(kContainer.AfterKanban.FileList))
			}
			break loop
		}
	} // end loop
	return nil
}

func receiveDirInfo(stream kanbanpb.SendAnything_SendToOtherDevicesServer, in *kanbanpb.SendContext, writableDirRoot string) (*outputDirInfo, error) {
	content := &kanbanpb.DirInfo{}
	if err := ptypes.UnmarshalAny(in.Context, content); err != nil {
		return nil, xerrors.Errorf("cant unmarshal any: %w", err)
	}
	// ちゃんととって来れているか確認
	LOG(content)
	parentDirPath := filepath.Join(writableDirRoot, content.RelParentPath)
	return downloadDir(stream, content, parentDirPath)
}

// sentDirOutputPath 受け取ったディレクトリの保存場所
func downloadDir(stream kanbanpb.SendAnything_SendToOtherDevicesServer, dirInfo *kanbanpb.DirInfo, parentDirPath string) (*outputDirInfo, error) {
	downloadDirInfo := &outputDirInfo{
		fileCnt: 0,
		files:   make([]*outputFileInfo, 0, dirInfo.FileCnt),
		dirName: dirInfo.DirName,
		path:    filepath.Join(parentDirPath, dirInfo.DirName),
	}
loop:
	for {
		in, err := stream.Recv()
		if err != nil {
			return nil, xerrors.Errorf("failed receiving kanban: %w", err)
		}
		switch in.Code {
		case kanbanpb.UploadRequestCode_SendingFile_Info:
			f, err := receiveFileInfo(stream, in, downloadDirInfo.path)
			if err != nil {
				return nil, xerrors.Errorf("failed receiving file info: %w", err)
			}
			downloadDirInfo.files = append(downloadDirInfo.files, f)
			downloadDirInfo.fileCnt++
		case kanbanpb.UploadRequestCode_EndOfSendingDir:
			break loop
		default:
			return nil, xerrors.Errorf("got unknown code %v", in.Code)
		}
	} // end loop
	return downloadDirInfo, nil
}

func receiveFileInfo(stream kanbanpb.SendAnything_SendToOtherDevicesServer, in *kanbanpb.SendContext, dirPath string) (*outputFileInfo, error) {
	content := &kanbanpb.FileInfo{}
	if err := ptypes.UnmarshalAny(in.Context, content); err != nil {
		return nil, xerrors.Errorf("cant unmarshal any: %w", err)
	}
	// ちゃんととって来れているか確認
	LOG(content)
	return downloadFile(stream, content, dirPath)
}

func downloadFile(stream kanbanpb.SendAnything_SendToOtherDevicesServer, fileInfo *kanbanpb.FileInfo, outputDirPath string) (*outputFileInfo, error) {
	outputFile := newReceivingFileBuilder(fileInfo, outputDirPath)
loop:
	for {
		in, err := stream.Recv()
		if err != nil {
			return nil, xerrors.Errorf("failed receiving kanban: %w", err)
		}

		switch in.Code {
		case kanbanpb.UploadRequestCode_SendingFile_CONT:
			err = receiveFileCont(in, outputFile)
			if err != nil {
				return nil, xerrors.Errorf("failed receiving file content: %w", err)
			}
		case kanbanpb.UploadRequestCode_SendingFile_EOF:
			if err = receiveEOF(in, fileInfo, outputFile, outputDirPath); err != nil { // TODO
				return nil, xerrors.Errorf("failed receiving EndOfFile: %w", err)
			}
			break loop
		default:
			return nil, xerrors.Errorf("got unknown code %v", in.Code)
		}
	} // end loop

	retInfo := &outputFileInfo{
		dirName: filepath.Dir(outputFile.relPath),
		path:    getAbsPath(outputDirPath, outputFile.relPath),
	}
	return retInfo, nil
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

func receiveFileCont(in *kanbanpb.SendContext, file *receivingFileBuilder) error {
	chunk := &kanbanpb.Chunk{}
	if err := ptypes.UnmarshalAny(in.Context, chunk); err != nil {
		return xerrors.Errorf("cant unmarshal any: %w", err)
	}

	// TODO validate関数作成
	if chunk.RefNum != file.expectRefNum {
		return xerrors.Errorf("sending data order is not correct. want:%d got:%d", file.expectRefNum, chunk.RefNum)
	}
	if filepath.Base(file.relPath) != chunk.Name {
		return xerrors.Errorf("sending data has no consistent. want:%s(not omitted: %s) got:%s", filepath.Base(file.relPath), file.relPath, chunk.Name)
	}

	file.enqueue(chunk.Context)
	return nil
}

func receiveEOF(in *kanbanpb.SendContext, fInfo *kanbanpb.FileInfo, file *receivingFileBuilder, outputDirPath string) error {
	fInfoEOF := &kanbanpb.FileInfo{}
	if err := ptypes.UnmarshalAny(in.Context, fInfoEOF); err != nil {
		return xerrors.Errorf("cant unmarshal any: %w", err)
	}

	outputDir := filepath.Join(outputDirPath, fInfoEOF.RelDir)
	filePath := filepath.Join(outputDir, fInfoEOF.Name)

	if f, err := os.Stat(outputDir); os.IsNotExist(err) || !f.IsDir() {
		if err = os.MkdirAll(outputDir, 0775); err != nil {
			log.Printf("[SendToOtherDevices server] cant create directory, %v", err)
		}
	}
	err := file.dequeue(outputDir) //TODO
	if err != nil {
		return xerrors.Errorf("failed output file: %w", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return xerrors.Errorf("cant open output path %s: %w", filePath, err)
	}
	defer f.Close()

	if string(fInfo.Hash) != getMD5(f) {
		os.Remove(f.Name())
		return xerrors.Errorf("Received file [%v] is broken", f.Name())
	}

	log.Printf("[SendToOtherDevices server] success to write file (%s)", filePath)
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

type eosCheck interface {
	ReceivedFileCnt() int
}

func receiveEOS(in *kanbanpb.SendContext, cntCheck eosCheck) error {
	eos := &kanbanpb.StreamInfo{}
	if err := ptypes.UnmarshalAny(in.Context, eos); err != nil {
		return xerrors.Errorf("cant unmarshal any: %w", err)
	}

	if eos.FileCount != int32(cntCheck.ReceivedFileCnt()) {
		return xerrors.Errorf("sent files count and received files count is NOT equal. got: %d, want: %d", cntCheck.ReceivedFileCnt(), eos.FileCount)
	}

	return nil
}

type outputItemInfo struct {
	files []*outputFileInfo
	dirs  []*outputDirInfo
}

type outputDirInfo struct {
	fileCnt int
	dirName string
	path    string
	files   []*outputFileInfo
}

type outputFileInfo struct {
	dirName string
	path    string
}

func newOutputItemInfo() *outputItemInfo {
	return &outputItemInfo{
		// make関数: スライス作成時、容量(cap)は長さと同じになる
		files: make([]*outputFileInfo, 0),
		dirs:  make([]*outputDirInfo, 0),
	}
}

func (i *outputItemInfo) list() []string {
	fList := make([]string, 0, len(i.dirs)+len(i.files))
	for _, d := range i.dirs {
		fList = append(fList, d.path)
	}
	for _, f := range i.files {
		fList = append(fList, f.path)
	}
	return fList
}

func (i *outputItemInfo) ReceivedFileCnt() int {
	cnt := 0
	for _, d := range i.dirs {
		cnt += d.fileCnt
	}
	cnt += len(i.files)
	return cnt
}

type receivingFileBuilder struct {
	relPath      string
	size         int64
	fileStack    []byte
	expectRefNum int32
}

func newReceivingFileBuilder(fInfo *kanbanpb.FileInfo, dirPath string) *receivingFileBuilder {
	content := make([]byte, 0, fInfo.Size)
	fb := &receivingFileBuilder{
		relPath:      filepath.Join(fInfo.RelDir, fInfo.Name),
		size:         fInfo.Size,
		fileStack:    content,
		expectRefNum: 0,
	}
	return fb
}

func (fb *receivingFileBuilder) enqueue(content []byte) {
	fb.fileStack = append(fb.fileStack, content...)
	fb.expectRefNum++
}

func (fb *receivingFileBuilder) dequeue(rootPath string) error {
	fPath := filepath.Join(rootPath, fb.relPath)
	f, err := os.Create(fPath)
	if err != nil {
		return xerrors.Errorf("cant open output path: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(fb.fileStack); err != nil {
		return xerrors.Errorf("cant write out file: %w", err)
	}
	fb.expectRefNum = 0
	return nil
}
