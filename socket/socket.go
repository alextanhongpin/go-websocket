package socket

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
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

func handleUnexpectedClose(err error) bool {
	return websocket.IsUnexpectedCloseError(err,
		websocket.CloseGoingAway,
		websocket.CloseAbnormalClosure)
}

func readDeadline() time.Time {
	return time.Now().Add(pongWait)
}

func writeDeadline() time.Time {
	return time.Now().Add(writeWait)
}
