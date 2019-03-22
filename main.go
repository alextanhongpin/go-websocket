package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/alextanhongpin/go-websocket/socket"
	"github.com/alextanhongpin/pkg/grace"
)

func main() {
	var shutdowns grace.Shutdowns
	hub := socket.NewHub()

	shutdowns.Append(hub.Start())

	fs := http.FileServer(http.Dir("public"))

	mux := http.NewServeMux()
	mux.Handle("/", fs)
	// mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
	//         fmt.Fprintf(w, `{"john"}`, nil)
	// })
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Perform authentication here? Use a middleware instead.
		hub.ServeWs(w, r)
	})
	shutdowns.Append(grace.New(mux, "8080"))

	quit := grace.Signal()
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shutdowns.Close(ctx)
	log.Println("graceful shutdown")
}
