package socket

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
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

func (c *Client) writePump(ctx context.Context) error {
	log.Println("write pump")
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-c.send:
			fmt.Println("received message", msg)
			c.conn.SetWriteDeadline(writeDeadline())
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return errors.New("send channel closed")
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return errors.Wrap(err, "write message failed")
			}
		case <-ticker.C:
			log.Println("writing ping")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return errors.Wrap(err, "write ping failed")
			}
		}
	}
}
