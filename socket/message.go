package socket

import "time"

// MessageType
type MessageType string

func (m MessageType) String() string {
	return string(m)
}

const (
	PingMessageType = MessageType("ping")
)

// TODO: Probably write anothe layer for message.

// Message is passed between websockets.
type Message struct {
	ID        string
	Data      map[string]interface{} `json:"data"`
	From      UserID                 `json:"from"`
	ConnID    ConnID                 `json:"-"`
	Timestamp time.Time              `json:"-"`
	To        UserID                 `json:"to"`
	Type      MessageType            `json:"type,omitempty"`
}
