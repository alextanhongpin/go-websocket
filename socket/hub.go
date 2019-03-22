package socket

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type Hub struct {
	sync.WaitGroup
	// Registered clients.
	clients   *Clients
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

func (h *Hub) Start() func(context.Context) {
	h.Add(1)
	go func() {
		defer h.Done()
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
			close(h.broadcast)
			h.Wait()
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

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		client.writePump()
	}()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(readDeadline())
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(readDeadline())
		return nil
	})
	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if handleUnexpectedClose(err) {
				log.Printf("error: %v", err)
			}
			break
		}
		msg.From = client.userID
		msg.ConnID = client.connID
		msg.Timestamp = time.Now()
		switch msg.Type {
		// Only ping to individual connection.
		case PingMessageType:
			_ = client.SendMessage(msg)
		default:
			h.broadcast <- msg
		}
	}
	h.clients.Remove(client)
	wg.Wait()
	log.Println("shutting down serverws")
}
