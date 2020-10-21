// Copyright (c) 2019-2020 Latona. All rights reserved.
package wsclient

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	wsScheme        = "ws"
	retryInterval   = 500
	maxNumOfMessage = 100
)

type stateType int

const (
	Ready stateType = iota
	Connected
	Disconnecting
	Closed
)

type WSClient interface {
	ConnectToServer() error
	SendMessage(string) error
	Reconnect()
	Close()
}

type wsClient struct {
	sync.Mutex
	conn          *websocket.Conn
	recvMessageCh chan string
	sendMessageCh chan string
	url           string
	state         stateType
}

func GetWebsocketClient(host string, port int, uri string) WSClient {
	hostname := host + ":" + strconv.Itoa(port)
	u := url.URL{Scheme: wsScheme, Host: hostname, Path: "/" + uri}

	ws := &wsClient{
		recvMessageCh: make(chan string, maxNumOfMessage),
		sendMessageCh: make(chan string, maxNumOfMessage),
		url:           u.String(),
		conn:          nil,
		state:         Ready,
	}
	go ws.loopConnectToServer(retryInterval)
	return ws
}

func (ws *wsClient) ConnectToServer() error {
	if ws.isState(Ready) {

		log.Printf("[websocket] connecting to %s", ws.url)
		conn, _, err := websocket.DefaultDialer.Dial(ws.url, nil)
		if err != nil {
			return fmt.Errorf("dial: %s", err.Error())
		}
		log.Printf("[websocket] success connecting to %s", ws.url)

		ws.conn = conn
		ws.setState(Connected)
		// recv message loop
		go func() {
			for {
				_, message, err := ws.conn.ReadMessage()
				if err != nil {
					log.Printf("[websocket] read error: %s", err.Error())
					// if close connection, dont reconnect
					if !ws.isState(Closed) {
						go ws.Reconnect()
					}
					return
				}
				if !ws.isState(Closed) {
					ws.recvMessageCh <- string(message)
				}
			}
		}()
		// send message loop
		go func() {
			for message := range ws.sendMessageCh {
				if err := ws.conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					log.Print("[websocket] write error: ", err)
					break
				}
			}
		}()
	}
	return nil
}

func (ws *wsClient) SendMessage(message string) error {
	if !ws.isState(Closed) {
		select {
		case ws.sendMessageCh <- message:
		default:
			return fmt.Errorf("sendMessage channel is full")
		}
		return nil
	}
	return fmt.Errorf("[websocket] connection is already closed")
}

func (ws *wsClient) loopConnectToServer(msInterval int) {
	cnt := 0
	for ws.isState(Ready) {
		if err := ws.ConnectToServer(); err == nil {
			break
		}
		// connection failed log
		cnt++
		log.Printf("[websocket] reconnection is failed (retry: %d, url: %s)", cnt, ws.url)
		time.Sleep(time.Millisecond * time.Duration(msInterval))
	}
}

func (ws *wsClient) Reconnect() {
	ws.disconnect()
	ws.setState(Ready)
	go ws.loopConnectToServer(retryInterval)
}

func (ws *wsClient) disconnect() {
	ws.setState(Disconnecting)
	if ws.isState(Connected) {
		// send close request
		_ = ws.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		// close connection
		ws.conn.Close()
	}
}

func (ws *wsClient) Close() {
	log.Printf("close conenction: %s", ws.url)
	ws.disconnect()
	ws.setState(Closed)
	close(ws.sendMessageCh)
	close(ws.recvMessageCh)
}

func (ws *wsClient) isState(state stateType) bool {
	ws.Lock()
	defer ws.Unlock()
	return ws.state == state
}

func (ws *wsClient) setState(state stateType) {
	ws.Lock()
	defer ws.Unlock()
	ws.state = state
}
