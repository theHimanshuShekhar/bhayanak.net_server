package models

import "time"

type Message struct {
	// Define the fields that match the expected JSON structure
	Type      string    `json:"type"`
	UserID    string    `json:"userID"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
