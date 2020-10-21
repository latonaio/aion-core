// Copyright (c) 2019-2020 Latona. All rights reserved.
package wsclient

import (
	"bitbucket.org/latonaio/aion-core/pkg/log"
	"context"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)

var upgrader = websocket.Upgrader{} // use default options

type resvType int

const (
	Success resvType = iota
	Failed
)

const (
	port    = 12120
	uri     = "test"
	host    = "localhost"
	urlBase = "ws://localhost:"
)

type WSServer struct {
	srv    *http.Server
	wg     *sync.WaitGroup
	recvCh chan resvType
}

func receiver(expectMsg string, responseCh chan resvType) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed:", err)
			responseCh <- Failed
			return
		}
		defer c.Close()
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Print("receive error: ", err)
			responseCh <- Failed
			return
		}
		if string(message) != expectMsg {
			log.Printf("cant get expect value (expect: %s, recv: %s)",
				expectMsg, string(message))
			responseCh <- Failed
		}
		responseCh <- Success
	}
}

func websocketResvServer(uri string, port int, expectMsg string) *WSServer {
	wg := &sync.WaitGroup{}
	recvCh := make(chan resvType)

	mux := http.NewServeMux()
	srv := &http.Server{
		Handler: mux,
		Addr:    ":" + strconv.Itoa(port),
	}

	mux.HandleFunc("/"+uri, receiver(expectMsg, recvCh))
	go func() {
		wg.Add(1)
		defer wg.Done()

		if err := srv.ListenAndServe(); err != nil {
			log.Print("ListenAndServe: ", err)
		}
	}()
	wss := &WSServer{
		srv:    srv,
		wg:     wg,
		recvCh: recvCh,
	}
	return wss
}

func (server *WSServer) stopServer() {
	if err := server.srv.Shutdown(context.TODO()); err != nil {
		log.Print("Shutdown: ", err)
	}
	server.wg.Wait()
}

func TestNormalConnectToServer(t *testing.T) {
	ws := &wsClient{url: urlBase + strconv.Itoa(port) + "/" + uri}
	server := websocketResvServer(uri, port, uri)
	defer server.stopServer()

	if err := ws.ConnectToServer(); err != nil {
		t.Error(err)
	}
}

func TestAbnormalConnectToServer(t *testing.T) {
	urlList := []string{
		urlBase + strconv.Itoa(8080) + "/" + uri,
		urlBase + strconv.Itoa(port) + "/" + "invalid",
	}
	for _, url := range urlList {
		t.Run("TestAbnormalConnectToServer", func(t *testing.T) {
			ws := &wsClient{url: url}
			server := websocketResvServer(uri, port, "")
			defer server.stopServer()

			if err := ws.ConnectToServer(); err == nil {
				t.Errorf("invalid result (url: %s)", url)
			}
		})
	}
}

func TestNormalSendMessage(t *testing.T) {
	expectMsg := "message"

	server := websocketResvServer(uri, port, expectMsg)
	defer server.stopServer()

	client := GetWebsocketClient(host, port, uri)
	defer client.Close()

	// testing method
	if err := client.SendMessage(expectMsg); err != nil {
		t.Fatal(err)
	}

	select {
	case res := <-server.recvCh:
		if res != Success {
			t.Errorf("cant get expect value")
		}
		break
	case <-time.After(time.Second * 2):
		t.Errorf("cant receive message")
	}
}

func TestAbnormal001SendMessage(t *testing.T) {
	expectMsg := "message"

	server := websocketResvServer(uri, port, expectMsg)
	defer server.stopServer()

	client := GetWebsocketClient(host, port, uri)
	defer client.Close()

	// testing method
	if err := client.SendMessage(""); err != nil {
		t.Fatal(err)
	}

	select {
	case res := <-server.recvCh:
		if res != Failed {
			t.Errorf("cant get expect value")
		}
		break
	case <-time.After(time.Second * 2):
		t.Errorf("cant receive message")
	}
}

func TestAbnormal002SendMessage(t *testing.T) {
	expectMsg := "message"

	server := websocketResvServer(uri, port, expectMsg)
	defer server.stopServer()

	client := GetWebsocketClient(host, port, uri)
	client.Close()

	// testing method
	if err := client.SendMessage(""); err == nil {
		t.Errorf("conenction is already closed, but success to send message")
	}
}

func TestNormal001Reconnect(t *testing.T) {
	expectMsg := "message"

	server := websocketResvServer(uri, port, expectMsg)
	defer server.stopServer()

	client := GetWebsocketClient(host, port, uri)
	defer client.Close()
	client.Reconnect()

	// testing method
	if err := client.SendMessage(expectMsg); err != nil {
		t.Fatal(err)
	}

	select {
	case res := <-server.recvCh:
		if res != Success {
			t.Errorf("cant get expect value")
		}
		break
	case <-time.After(time.Second * 2):
		t.Errorf("cant receive message")
	}
}

func TestNormal002Reconnect(t *testing.T) {
	expectMsg := "message"

	server := websocketResvServer(uri, port, expectMsg)
	client := GetWebsocketClient(host, port, uri)
	defer client.Close()

	// check to execute reconnect when restart server
	server.stopServer()

	// testing method
	if err := client.SendMessage(expectMsg); err != nil {
		t.Fatal(err)
	}

	server = websocketResvServer(uri, port, expectMsg)
	defer server.stopServer()

	select {
	case res := <-server.recvCh:
		if res != Success {
			t.Errorf("cant get expect value")
		}
		break
	case <-time.After(time.Second * 2):
		t.Errorf("cant receive message")
	}
}
