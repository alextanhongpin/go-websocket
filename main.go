package main

import (
	"log"
	"net/http"

	"github.com/alextanhongpin/go-websocket/socket"
)

func main() {
	hub := socket.NewHub()
	shutdown := hub.Start()
	defer shutdown()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Perform authentication here? Use a middleware instead.
		socket.ServeWs(hub, w, r)
	})
	log.Println("listening to port *:8080. press ctrl + c to cancel")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
