package models

import (
	"github.com/gorilla/websocket"
)

type Connection struct {
	Conn     *websocket.Conn
	User     string    // Username
	ChatRoom *ChatRoom // Channel
}
