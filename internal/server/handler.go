package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/theHimanshuShekhar/bhayanak.net/pkg/logger"
	"github.com/theHimanshuShekhar/bhayanak.net/pkg/models"
	"github.com/theHimanshuShekhar/bhayanak.net/pkg/services"
)

// Upgrader specifies the parameters for upgrading an HTTP connection to a WebSocket connection.
var upgrader = websocket.Upgrader{}

// HandleConnections upgrades HTTP requests to WebSocket connections and manages communication
func HandleConnections(w http.ResponseWriter, r *http.Request) {

	logger.Info(fmt.Sprintf("New connection request from %s", r.RemoteAddr))

	// Upgrade initial GET request to a WebSocket connection
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade to a WebSocket connection: ", err)
		w.Write([]byte("Failed to upgrade to a WebSocket connection. Please check your client.\nError: " + err.Error()))
		return
	}
	defer conn.Close()

	// create a new connection with channel as general
	currentConnection := models.Connection{
		Conn: conn,
		User: uuid.NewString(),
	}

	// By default, join the general channel
	services.JoinChatRoom(&currentConnection, "general")

	// Defer the removal of the connection after the user disconnects
	defer func() {
		services.LeaveChatRoom(&currentConnection)
	}()

	// Handle incoming messages from the WebSocket connection
	for {
		// Read JSON message from the WebSocket connection using ReadJSON
		_, message, err := conn.ReadMessage()
		if err != nil {
			// if error is close 1005, it means the connection is closed
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				logger.Info("Connection closed by the client")
				return
			}
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				logger.Info("Connection closed by the client due to no status or abnormal closure")
				return
			}
			logger.Error("Failed to read message from the WebSocket connection: ", err)
			return
		}

		// Unmarshal JSON message into the Message struct
		var msg models.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Error("Failed to unmarshal JSON message: ", err)
			continue
		}

		services.BroadcastMessage(&currentConnection, msg)
	}
}
