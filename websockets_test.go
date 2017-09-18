package websockets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/gurupras/go-stoppable-net-listener"
	"github.com/stretchr/testify/require"
)

func setupWS() (http.Server, *stoppablenetlistener.StoppableNetListener, *WebsocketServer) {
	router := mux.NewRouter()
	ws := NewServer(router)
	server := http.Server{}
	server.Handler = router
	snl, err := stoppablenetlistener.New(51221)
	if err != nil {
		log.Fatalf("Failed to listen to port 51221: %v", err)
	}
	return server, snl, ws
}

func stopServer(snl *stoppablenetlistener.StoppableNetListener) {
	snl.Stop()
}

func TestMsgPack(t *testing.T) {
	require := require.New(t)
	data := []int{1, 2, 3}
	m := WebsocketMessage{"test", data}

	b, err := msgpack.Marshal(&m)
	require.Nil(err)

	var got WebsocketMessage
	err = msgpack.Unmarshal(b, &got)
	require.Nil(err)

	expectedStr, err := json.Marshal(&m)
	require.Nil(err)

	gotStr, err := json.Marshal(&got)
	require.Nil(err)

	require.JSONEq(string(expectedStr), string(gotStr))
}

func TestListener(t *testing.T) {
	require := require.New(t)
	server, snl, ws := setupWS()

	wg := sync.WaitGroup{}
	wg.Add(1)
	ws.On("test", func(w *WebsocketClient, data interface{}) {
		require.Nil(data)
		wg.Done()
	})
	go server.Serve(snl)
	defer stopServer(snl)

	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:51221",
		Path:   "/ws",
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	require.Nil(err)
	require.NotNil(c)
	client := NewClient(c)
	err = client.Emit("test", nil)
	require.Nil(err)
	wg.Wait()
}

func TestMultipleListener(t *testing.T) {
	require := require.New(t)
	server, snl, ws := setupWS()

	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		ws.On("test", func(w *WebsocketClient, data interface{}) {
			wg.Done()
		})
	}
	go server.Serve(snl)
	defer stopServer(snl)

	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:51221",
		Path:   "/ws",
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	require.Nil(err)
	require.NotNil(c)
	client := NewClient(c)
	client.Emit("test", nil)
	require.Nil(err)
	wg.Wait()
}

func TestEmit(t *testing.T) {
	require := require.New(t)
	server, snl, ws := setupWS()

	wg := sync.WaitGroup{}
	wg.Add(1)
	counter := 0
	ws.On("ping", func(w *WebsocketClient, data interface{}) {
		log.Debugf("server ping")
		expected := counter
		_ = expected
		got := data
		_ = got
		log.Debugf("Expected '%v'  Got '%v'", expected, got)
		log.Debugf("Attempting to emit pong on socket")
		w.Emit("pong", counter)
	})
	go server.Serve(snl)
	defer stopServer(snl)

	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:51221",
		Path:   "/ws",
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	require.Nil(err)
	require.NotNil(c)
	client := NewClient(c)
	go client.ProcessMessages()
	client.On("pong", func(w *WebsocketClient, data interface{}) {
		log.Debugf("client pong")
		expected := fmt.Sprintf("%v", counter)
		_ = expected
		counter++
		if counter == 10 {
			wg.Done()
		} else {
			client.Emit("ping", counter)
		}
	})
	client.Emit("ping", counter)
	require.Nil(err)
	wg.Wait()
}

func TestClose(t *testing.T) {
	require := require.New(t)
	server, snl, _ := setupWS()

	go server.Serve(snl)
	defer stopServer(snl)

	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:51221",
		Path:   "/ws",
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	require.Nil(err)
	require.NotNil(c)
	client := NewClient(c)
	err = client.Close()
	require.Nil(err)
}
