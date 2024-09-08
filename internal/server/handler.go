package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader specifies the parameters for upgrading an HTTP connection to a WebSocket connection
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

// HandleConnections upgrades HTTP requests to WebSocket connections and manages communication
func HandleConnections(w http.ResponseWriter, r *http.Request) {
    // Upgrade initial GET request to a WebSocket connection
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Failed to upgrade to WebSocket: %v", err)
        return
    }
    defer conn.Close()

    for {
        // Read message from WebSocket client
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Error reading message: %v", err)
            break
        }

        // Echo message back to the client
        if err = conn.WriteMessage(messageType, message); err != nil {
            log.Printf("Error writing message: %v", err)
            break
        }
    }
}
