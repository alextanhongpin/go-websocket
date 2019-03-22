package socket

import (
	"log"
	"sync"
)

const (
	PingMsg = "ping"
)

type Hub struct {
	sync.WaitGroup
	// Registered clients.
	clients    map[UserID]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	quit       chan struct{}
	broadcast  chan Message
}

func NewHub() *Hub {
	return &Hub{
		quit:       make(chan struct{}),
		clients:    make(map[UserID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message),
	}
}

func (h *Hub) Start() func() {
	h.Add(1)
	go func() {
		defer func() {
			close(h.register)
			close(h.unregister)
			close(h.broadcast)
			h.Done()
		}()
		for {
			select {
			case <-h.quit:
				return
			case client, ok := <-h.register:
				if !ok {
					return
				}
				uid := client.userID
				if _, exist := h.clients[uid]; !exist {
					h.clients[uid] = make(map[*Client]bool)
				}
				h.clients[uid][client] = true
				log.Println("registered", uid)
			case client, ok := <-h.unregister:
				if !ok {
					return
				}
				if clients, exist := h.clients[client.userID]; exist {
					for c := range clients {
						if c.socketID == client.socketID {
							delete(clients, c)
							log.Println("unregistered", client.userID)
							if len(clients) == 0 {
								delete(h.clients, client.userID)
								close(client.send)
								log.Println("cleared user")
							}
							break
						}
					}
				}
			case msg, ok := <-h.broadcast:
				// Decide on which client to send the message
				// to.
				if !ok {
					return
				}
				log.Println("got message", msg)
				switch msg.Type {
				case PingMsg:
					userID := msg.From
					log.Println("received ping message")
					clients, exist := h.clients[userID]
					if !exist {
						delete(h.clients, userID)
						continue
					}
					// Should only ping yourself.
					for client := range clients {
						if client.socketID == msg.SocketID {
							client.send <- msg
							break
						}
					}
				default:
				}
			}
		}
	}()
	return func() {
		close(h.quit)
		h.Wait()
	}
}
