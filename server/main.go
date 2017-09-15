package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gurupras/go-stoppable-net-listener"
	"github.com/homesound/simple-websockets"
	log "github.com/sirupsen/logrus"
)

func serveHome(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadFile("./static/html/index.html")
	if err != nil {
		log.Fatalf("Failed to serve index.html: %v", err)
	}
	w.Write(b)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", serveHome)
	router.PathPrefix("/static").Handler(http.FileServer(http.Dir("./")))
	ws := websockets.NewWebsocketServer(router)
	server := http.Server{}
	server.Handler = router
	snl, err := stoppablenetlistener.New(51221)
	if err != nil {
		log.Fatalf("Failed to listen to port 51221: %v", err)
	}
	ws.On("echo", func(w *websockets.WebsocketClient, msg []byte) error {
		log.Infof("Received: %v", string(msg))
		w.Emit("echo", []byte(fmt.Sprintf("Successfully received: %v", string(msg))))
		return nil
	})
	log.Infof("Running server on port: 51221")
	server.Serve(snl)
}
