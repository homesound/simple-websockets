package websockets

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"

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

func TestListener(t *testing.T) {
	require := require.New(t)
	server, snl, ws := setupWS()

	wg := sync.WaitGroup{}
	wg.Add(1)
	ws.On("test", func(w *WebsocketClient, msg []byte) {
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
	msg := make(map[string]interface{})
	msg[WS_EVENT_KEY] = "test"
	err = c.WriteJSON(msg)
	require.Nil(err)
	wg.Wait()
}

func TestMultipleListener(t *testing.T) {
	require := require.New(t)
	server, snl, ws := setupWS()

	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		ws.On("test", func(w *WebsocketClient, msg []byte) {
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
	msg := make(map[string]interface{})
	msg[WS_EVENT_KEY] = "test"
	err = c.WriteJSON(msg)
	require.Nil(err)
	wg.Wait()
}

func TestEmit(t *testing.T) {
	require := require.New(t)
	server, snl, ws := setupWS()

	wg := sync.WaitGroup{}
	wg.Add(1)
	counter := 0
	ws.On("ping", func(w *WebsocketClient, msg []byte) {
		log.Debugf("server ping")
		expected := fmt.Sprintf("%v", counter)
		_ = expected
		log.Debugf("Attempting to emit pong on socket")
		w.Emit("pong", []byte(fmt.Sprintf("%v", counter)))
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
	client.On("pong", func(w *WebsocketClient, msg []byte) {
		log.Debugf("client pong")
		expected := fmt.Sprintf("%v", counter)
		_ = expected
		counter++
		if counter == 10 {
			wg.Done()
		} else {
			client.Emit("ping", fmt.Sprintf("%v", counter))
		}
	})
	client.Emit("ping", fmt.Sprintf("%v", counter))
	require.Nil(err)
	wg.Wait()
}
