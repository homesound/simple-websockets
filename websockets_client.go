package websockets

import (
	"encoding/json"
	"strings"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type WebsocketClient struct {
	*websocket.Conn
	listenerMap map[string][]WebsocketListener
}

func NewClient(w *websocket.Conn) *WebsocketClient {
	wc := WebsocketClient{}
	wc.Conn = w
	wc.listenerMap = make(map[string][]WebsocketListener)
	return &wc
}

type WebsocketListener func(w *WebsocketClient, msg []byte)

func (wc *WebsocketClient) Emit(event string, data interface{}) error {
	message := make(map[string]interface{})
	switch data.(type) {
	case string:
		// This is a string. Encapsulate it in a special map
		// which the JS knows to decapsulate
		message[WS_TYPE_KEY] = WS_STRING_MESSAGE_KEY
		message[WS_STRING_MESSAGE_KEY] = data
	default:
		b, err := json.Marshal(data)
		if err != nil {
			log.Errorf("Failed to marshal emit's data: %v", err)
			return err
		}
		_ = json.Unmarshal(b, &message)
	}
	message[WS_EVENT_KEY] = event
	log.Debugf("Emitting: %v", message)
	return wc.WriteJSON(message)
}

func (wc *WebsocketClient) On(event string, fn WebsocketListener) {
	if _, ok := wc.listenerMap[event]; !ok {
		wc.listenerMap[event] = make([]WebsocketListener, 0)
	}
	wc.listenerMap[event] = append(wc.listenerMap[event], fn)
}

func (wc *WebsocketClient) ProcessMessages() {
	for {
		_, message, err := wc.ReadMessage()
		if err != nil {
			log.Errorf("Failed to read message from websocket: %v", err)
			continue
		}
		log.Debugf("Received message: %v", strings.TrimSpace(string(message)))

		msg := make(map[string]interface{})
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Errorf("Failed to unmarshal message from websocket: %v", err)
			break
		}
		_event, ok := msg[WS_EVENT_KEY]
		if !ok {
			log.Errorf("No event specified: %v", message)
			continue
		}
		event := _event.(string)
		if listeners, ok := wc.listenerMap[event]; !ok {
			log.Warnf("Received unknown event: %v", event)
			continue
		} else {
			for _, listener := range listeners {
				listener(wc, message)
			}
		}
	}
}
