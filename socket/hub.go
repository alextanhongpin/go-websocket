package socket

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type Hub struct {
	// Registered clients.
	clients *Clients

	wg        sync.WaitGroup
	once      sync.Once
	broadcast chan Message
	// TODO: store userID:connID=hostID in redis, map connID to *websocket.Conn.
	// Register redis publish subscribe to existing hostID.
	// When there's a new message, publish it to the userID
	// Distinguish between distribute events and local (e.g. ping is local to the current connection only)
}

func NewHub() *Hub {
	return &Hub{
		clients:   NewClients(),
		broadcast: make(chan Message),
	}
}

func (h *Hub) close() {
	h.once.Do(func() {
		close(h.broadcast)
	})
	h.wg.Wait()
}

func (h *Hub) Start() func(context.Context) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		for {
			select {
			case msg, ok := <-h.broadcast:
				// Decide on which client to send the message
				// to.
				if !ok {
					return
				}
				log.Println("got message", msg)
				switch msg.Type {
				default:
					// Send to all?
				}
			}
		}
	}()
	return func(ctx context.Context) {
		done := make(chan struct{})
		go func() {
			h.close()
			close(done)
		}()
		select {
		case <-ctx.Done():
			return
		case <-done:
			log.Println("websocket graceful shutdown")
			return
		}
	}
}

func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()
	// Extract as middleware.
	// token := r.URL.Query().Get("token")
	// Validates the token string.
	// _ = token
	// MessageType: TextMessage, BinaryMessage, CloseMessage, PingMessage, PongMessage
	// if err := conn.WriteMessage(
	//         websocket.CloseMessage,
	//         websocket.FormatCloseMessage(websocket.CloseProtocolError, "unauthorized attempt"),
	// ); err != nil {
	//         log.Fatal(err)
	//         return
	// }
	client := NewClient("1", conn)
	h.clients.Add(client)
	log.Println("added client", client.connID, client.userID)

	shutdown := client.writePump()
	// Blocking.
	client.readPump(h.broadcast)

	// No more reads, remove the client first before shutting down the send
	// channel to avoid sending to closed channel.
	h.clients.Remove(client)
	log.Println("removed client", client.connID, client.userID)

	// Shutdown within 5 seconds, else cancel.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Perform shutdown with cancellation.
	shutdown(ctx)
	log.Println("shutdown serveWs")
}
