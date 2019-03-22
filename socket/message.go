package socket

import "time"

// MessageType
type MessageType string

const (
	PingMessageType = MessageType("ping")
)

type Message struct {
	ID        string
	Data      map[string]interface{} `json:"data"`
	From      UserID                 `json:"from"`
	ConnID    ConnID                 `json:"-"`
	Timestamp time.Time              `json:"-"`
	To        UserID                 `json:"to"`
	Type      MessageType            `json:"type,omitempty"`
}
