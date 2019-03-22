package socket

import "github.com/rs/xid"

type ConnID string

func (c ConnID) String() string {
	return string(c)
}

func NewConnID() ConnID {
	// TODO: Check better implementation (ordered uuid, ksuid)
	return ConnID(xid.New().String())
}
