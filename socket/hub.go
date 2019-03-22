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
	clients *Clients

	quit chan struct{}

	broadcast chan Message
}

func NewHub() *Hub {
	return &Hub{
		quit:      make(chan struct{}),
		clients:   NewClients(),
		broadcast: make(chan Message),
	}
}

func (h *Hub) Start() func(context.Context) {
	h.Add(1)
	go func() {
		defer func() {
			close(h.broadcast)
			h.Done()
		}()
		for {
			select {
			case <-h.quit:
				return
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
			close(h.quit)
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
	undo := h.clients.Add(client)
	defer undo()

	ctx := r.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		err := client.writePump(ctx)
		if err != nil {
			log.Println(err)
		}
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
}
