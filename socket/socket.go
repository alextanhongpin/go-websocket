package socket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"
)

// User is already logged in.
// User is connected to websocket.
// User sends token to the server.
// The server validates the token and obtain the user id.
// The server creates a new id and assigns the socket id to the user id.
// The user id can have several socket id, since the user can be connected to multiple devices. There is a cap of 10 devices per user.

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SocketID string

func NewSocketID() SocketID {
	return SocketID(xid.New().String())
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	// Extract as middleware.
	// token := r.URL.Query().Get("token")
	// Validates the token string.
	// _ = token
	// MessageType: TextMessage, BinaryMessage, CloseMessage, PingMessage, PongMessage
	if err := conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseProtocolError, "unauthorized attempt"),
	); err != nil {
		log.Fatal(err)
		return
	}
	client := NewClient("1", hub, conn)
	log.Println("registering client")
	client.hub.register <- client
	go client.writePump()
	go client.readPump()
}
