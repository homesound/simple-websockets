package main

import (
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
	log.SetLevel(log.DebugLevel)
	router := mux.NewRouter()
	router.HandleFunc("/", serveHome)
	router.PathPrefix("/static").Handler(http.FileServer(http.Dir("./")))
	ws := websockets.NewServer(router)
	server := http.Server{}
	server.Handler = router
	snl, err := stoppablenetlistener.New(51221)
	if err != nil {
		log.Fatalf("Failed to listen to port 51221: %v", err)
	}
	ws.On("echo", func(w *websockets.WebsocketClient, data interface{}) {
		log.Infof("Received: %v", data)
		w.Emit("echo", data)
	})
	log.Infof("Running server on port: 51221")
	server.Serve(snl)
}
