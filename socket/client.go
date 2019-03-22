package socket

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub       *Hub
	socketID  SocketID
	userID    UserID
	conn      *websocket.Conn
	send      chan Message
	createdAt time.Time
	// List of peers that can communicate with the user.
	// allowed []UserID
}

func NewClient(id string, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		userID:    UserID(id),
		hub:       hub,
		conn:      conn,
		socketID:  NewSocketID(),
		send:      make(chan Message),
		createdAt: time.Now(),
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		log.Println("read pump closed")
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			fmt.Println("error reading json", err)
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		msg.From = c.userID
		msg.SocketID = c.socketID
		msg.Timestamp = time.Now()
		fmt.Println("got msg:", msg)
		c.hub.broadcast <- msg
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		log.Println("write pump closed")
	}()
	for {
		select {
		case msg, ok := <-c.send:
			fmt.Println("received message", msg)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				log.Println("write json failed:", err)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
		case <-ticker.C:
			log.Println("writing ping")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("write message failed", err)
				return
			}
		}
	}

}
