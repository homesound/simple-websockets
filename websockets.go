package websockets

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type WebsocketServer struct {
	Clients     map[*WebsocketClient]struct{}
	upgrader    websocket.Upgrader
	listenerMap map[string][]WebsocketListener
}

func NewServer(router *mux.Router) *WebsocketServer {
	ws := WebsocketServer{}
	ws.Clients = make(map[*WebsocketClient]struct{})
	ws.upgrader = websocket.Upgrader{}
	ws.listenerMap = make(map[string][]WebsocketListener)

	router.HandleFunc("/ws", ws.handleConnections)
	return &ws
}

func (ws *WebsocketServer) handleConnections(wr http.ResponseWriter, r *http.Request) {
	w, err := ws.upgrader.Upgrade(wr, r, nil)
	if err != nil {
		log.Fatalf("Failed to upgrade to websockets: %v", err)
	}
	client := NewClient(w)
	defer w.Close()
	ws.Clients[client] = struct{}{}
	ws.addListeners(client)
	client.ProcessMessages()
}

func (ws *WebsocketServer) On(event string, fn WebsocketListener) {
	if _, ok := ws.listenerMap[event]; !ok {
		ws.listenerMap[event] = make([]WebsocketListener, 0)
	}
	ws.listenerMap[event] = append(ws.listenerMap[event], fn)
	for client, _ := range ws.Clients {
		client.On(event, fn)
	}
}

func (ws *WebsocketServer) addListeners(wc *WebsocketClient) {
	for event, functions := range ws.listenerMap {
		for _, fn := range functions {
			wc.On(event, fn)
		}
	}
}
