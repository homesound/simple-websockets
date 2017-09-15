package websockets

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
)

type WebsocketClient struct {
	*websocket.Conn
	listenerMap map[string][]WebsocketListener
}

type WebsocketMessage struct {
	Event string      `msgpack:"event"`
	Data  interface{} `msgpack:"data"`
}

func NewClient(w *websocket.Conn) *WebsocketClient {
	wc := WebsocketClient{}
	wc.Conn = w
	wc.listenerMap = make(map[string][]WebsocketListener)
	return &wc
}

type WebsocketListener func(w *WebsocketClient, data interface{})

func (wc *WebsocketClient) Emit(event string, data interface{}) error {
	msg := WebsocketMessage{event, data}

	b, err := msgpack.Marshal(&msg)
	if err != nil {
		log.Errorf("Failed to marshal: %v", err)
		return err
	}
	log.Debugf("Emitting event=%v data=%v len=%v", event, data, len(b))
	if err := wc.WriteMessage(websocket.BinaryMessage, b); err != nil {
		log.Errorf("Failed to write message: %v", err)
		return err
	}
	return nil
}

func (wc *WebsocketClient) On(event string, fn WebsocketListener) {
	if _, ok := wc.listenerMap[event]; !ok {
		wc.listenerMap[event] = make([]WebsocketListener, 0)
	}
	wc.listenerMap[event] = append(wc.listenerMap[event], fn)
}

func (wc *WebsocketClient) ProcessMessages() {
	for {
		t, message, err := wc.ReadMessage()
		if err != nil {
			log.Errorf("Failed to read message from websocket: %v", err)
			continue
		}
		log.Debugf("Received message type=%v message=%v", t, message)
		msg := WebsocketMessage{}
		msgpack.Unmarshal(message, &msg)
		event := msg.Event
		log.Debugf("event=%v data=%v", event, msg.Data)
		if listeners, ok := wc.listenerMap[event]; !ok {
			log.Warnf("Received unknown event: %v", event)
			continue
		} else {
			for _, listener := range listeners {
				listener(wc, msg.Data)
			}
		}

	}
}
