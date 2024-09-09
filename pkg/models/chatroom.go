package models

import "sync"

type ChatRoom struct {
	Name           string
	ConnectedUsers []*Connection
	Mu             sync.Mutex // Mutex to ensure safe concurrent access
}
