package socket

import (
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	connID    ConnID
	userID    UserID
	conn      *websocket.Conn
	send      chan Message
	createdAt time.Time
	// List of peers that can communicate with the user.
	// allowed []UserID
}

func (c *Client) SendMessage(msg Message) bool {
	select {
	case c.send <- msg:
		return true
	case <-time.After(5 * time.Second):
		return false
	}
}

func NewClient(id string, conn *websocket.Conn) *Client {
	return &Client{
		userID:    UserID(id),
		conn:      conn,
		connID:    NewConnID(),
		send:      make(chan Message),
		createdAt: time.Now(),
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(writeDeadline())
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
