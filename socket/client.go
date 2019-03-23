package socket

import (
	"log"
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

func (c *Client) readPump(broadcast chan<- Message) {
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(readDeadline())
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(readDeadline())
		return nil
	})
	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			if handleUnexpectedClose(err) {
				log.Printf("error: %v", err)
			}
			break
		}
		msg.From = c.userID
		msg.ConnID = c.connID
		msg.Timestamp = time.Now()
		switch msg.Type {
		// Only ping to individual connection.
		case PingMessageType:
			_ = c.SendMessage(msg)
		default:
			broadcast <- msg
		}
	}
}
