// Copyright (c) 2019-2020 Latona. All rights reserved.
package sftp

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
)

type sftpState int

const (
	Initialize sftpState = iota
	Ready
	Disconnect
	Stop
)

type FilePathContainer struct {
	srcPath string
	dstPath string
}

var stopSendRequest = FilePathContainer{"", ""}

const (
	sshTimeoutDuration time.Duration = 3 * time.Second
	sshRetryCount                    = 3
)

type SFTPClient struct {
	sync.Mutex
	sshConn    ssh.Conn
	client     *sftp.Client
	config     *ssh.ClientConfig
	url        string
	sendPathCh chan FilePathContainer
	status     sftpState
}

func NewSFTPClient(user string, password string, host string, port int) *SFTPClient {
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout: sshTimeoutDuration,
	}
	config.SetDefaults()
	url := host + ":" + strconv.Itoa(port)

	var sshConn ssh.Conn
	ss := &SFTPClient{
		sshConn:    sshConn,
		client:     nil,
		config:     config,
		url:        url,
		status:     Initialize,
		sendPathCh: make(chan FilePathContainer, 100),
	}
	go ss.createSFTPConnection()

	return ss
}

func (ss *SFTPClient) createSFTPConnection() error {
	if !(ss.status == Initialize || ss.status == Disconnect) {
		return nil
	}
	ss.Lock()
	defer ss.Unlock()
	// create ssh connection
	if err := retry.Do(
		func() error {
			sshConn, err := ssh.Dial("tcp", ss.url, ss.config)
			if err != nil {
				log.Print("[sftp] ssh connection is failed: ", ss.url, err)
				return err
			}
			// create sftp client
			client, err := sftp.NewClient(sshConn)
			if err != nil {
				log.Print("[sftp] ssh connection is failed: ", ss.url, err)
				return err
			}

			log.Printf("[sftp] success to connect : %s", ss.url)
			ss.client = client
			ss.sshConn = sshConn
			ss.status = Ready

			go func() {
				// wait to close connection
				_ = ss.client.Wait()

				// set closed status
				ss.Lock()
				defer ss.Unlock()
				if ss.status != Stop {
					log.Printf("[sftp] ssh connection disconnected: %s", ss.url)
					ss.status = Disconnect
					ss.sendPathCh <- stopSendRequest
				}
			}()

			// send loop
			go ss.waitSendFile()
			return nil
		},
		retry.Attempts(sshRetryCount),
	); err != nil {
		return err
	}
	return nil
}

func (ss *SFTPClient) waitSendFile() {
	for filePath := range ss.sendPathCh {
		// if status is stop, break this loop
		if filePath == stopSendRequest {
			log.Print("[sftp sender] sender function is closed.")
			break
		}
		// send files
		if err := ss.sendFile(filePath); err != nil {
			log.Print(err)
		}
	}
}

func (ss *SFTPClient) sendFile(filePath FilePathContainer) error {
	// open source file
	f, err := ioutil.ReadFile(filePath.srcPath)
	if err != nil {
		return fmt.Errorf("[SFTP Error] cant read file from local: " + filePath.srcPath)
	}

	// create remote file
	file, err := ss.client.Create(filePath.dstPath)
	if err != nil {
		return fmt.Errorf("[SFTP Error] cant create file to remote: " + filePath.dstPath)
	}
	defer file.Close()

	// write source file to remote
	_, err = file.Write(f)
	if err != nil {
		return fmt.Errorf("[sftp] failed to send file (src: %s, dst:%s)", filePath.srcPath, filePath.dstPath)
	}
	log.Printf("[sftp] success to send file (src: %s, dst:%s)", filePath.srcPath, filePath.dstPath)
	return nil
}

func (ss *SFTPClient) SetToSendPathCh(src string, dst string) error {
	if ss.status != Stop {
		// try to connect at remote device
		go ss.createSFTPConnection()
		select {
		case ss.sendPathCh <- FilePathContainer{src, dst}:
			break
		default:
			return fmt.Errorf("[SFTP Error] file is dropped: %s", src)
		}
	}
	return nil
}

func (ss *SFTPClient) Close() {
	ss.Lock()
	defer ss.Unlock()

	if ss.status != Stop {
		log.Printf("[sftp] close connection from %s", ss.url)
		if ss.status != Initialize {
			if err := ss.client.Close(); err != nil {
				log.Print("[sftp] cause error when close connection: ", err)
			}
			if err := ss.sshConn.Close(); err != nil {
				log.Print("[sftp] cause error when close connection: ", err)
			}
		}
		close(ss.sendPathCh)
		ss.status = Stop
	}
}
