package socket

import "time"

type Message struct {
	Data      map[string]interface{} `json:"data"`
	From      UserID                 `json:"from"`
	SocketID  SocketID               `json:"-"`
	Timestamp time.Time              `json:"-"`
	To        UserID                 `json:"to"`
	Type      string                 `json:"type,omitempty"`
}
