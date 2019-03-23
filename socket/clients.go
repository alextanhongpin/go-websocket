package socket

import (
	"log"
	"sync"
	"time"
)

type Clients struct {
	sync.RWMutex
	value map[UserID]map[*Client]struct{}
}

func NewClients() *Clients {
	return &Clients{
		value: make(map[UserID]map[*Client]struct{}),
	}
}

func (c *Clients) WithUserID(userID UserID) ([]*Client, bool) {
	c.RLock()
	clients, exist := c.value[userID]
	c.RUnlock()
	result := make([]*Client, len(clients))
	var i int
	for client := range clients {
		result[i] = client
		i++
	}
	return result, exist
}

func (c *Clients) Add(client *Client) {
	c.Lock()
	if _, exist := c.value[client.userID]; !exist {
		c.value[client.userID] = make(map[*Client]struct{})
	}
	c.value[client.userID][client] = struct{}{}
	c.Unlock()
}

func (c *Clients) Remove(client *Client) {
	c.Lock()
	if clients, exist := c.value[client.userID]; exist {
		log.Printf("userID=%s connID=%s duration=%s\n",
			client.userID,
			client.connID,
			time.Since(client.createdAt).String())
		delete(clients, client)
		if len(clients) == 0 {
			delete(c.value, client.userID)
		}
	}
	c.Unlock()
}
