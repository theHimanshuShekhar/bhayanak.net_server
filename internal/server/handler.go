package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/theHimanshuShekhar/bchat-server/pkg/logger"
)

type Channel struct {
	name           string
	connectedUsers []*Connection
	mu             sync.Mutex // Mutex to ensure safe concurrent access
}

type Connection struct {
	conn    *websocket.Conn
	user    string   // Username
	channel *Channel // Channel
}

// List of all the channels
var channels []Channel = []Channel{
	{
		name:           "general",
		connectedUsers: []*Connection{},
	},
	{
		name:           "fun",
		connectedUsers: []*Connection{},
	},
	{
		name:           "pol",
		connectedUsers: []*Connection{},
	},
	{
		name:           "tech",
		connectedUsers: []*Connection{},
	},
}

type Message struct {
	// Define the fields that match the expected JSON structure
	Type      string    `json:"type"`
	UserID    string    `json:"userID"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Upgrader specifies the parameters for upgrading an HTTP connection to a WebSocket connection.
var upgrader = websocket.Upgrader{}

func LogChannelUsers(channel *Channel) {
	logger.Info(fmt.Sprintf("Channel: %s | Connected Users: %d", channel.name, len(channel.connectedUsers)))

	// Use a slice to collect the users' IDs
	var users []string
	// Lock the channel before accessing the users if necessary
	channel.mu.Lock()
	for _, connection := range channel.connectedUsers {
		users = append(users, "["+connection.user+"]")
	}
	channel.mu.Unlock()
	// Join the users into a single string with space as a separator
	logger.Info(joinUsers(users))
}

// Helper function to join users into a string
func joinUsers(users []string) string {
	if len(users) == 0 {
		return "No users connected."
	}
	return "Users: " + strings.Join(users, " ")
}

// HandleConnections upgrades HTTP requests to WebSocket connections and manages communication
func HandleConnections(w http.ResponseWriter, r *http.Request) {

	logger.Info(fmt.Sprintf("New connection request from %s", r.RemoteAddr))

	// Upgrade initial GET request to a WebSocket connection
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade to a WebSocket connection: ", err)
		return
	}
	defer conn.Close()

	// create a new connection with channel as general
	newConnection := Connection{
		conn: conn,
		user: uuid.NewString(),
	}

	// create a new connection with channel as reference to general channel
	newConnection.channel = &channels[0]

	// Add the new connection to the channel safely
	newConnection.channel.mu.Lock()
	newConnection.channel.connectedUsers = append(newConnection.channel.connectedUsers, &newConnection)
	newConnection.channel.mu.Unlock()

	logger.Info(fmt.Sprintf("User %s joined channel: %s", newConnection.user, newConnection.channel.name))

	LogChannelUsers(newConnection.channel)

	// Defer the removal of the connection after the user disconnects
	defer func() {
		// Remove the connection from the channel safely
		newConnection.channel.mu.Lock()
		for i, connection := range newConnection.channel.connectedUsers {
			if connection.conn == newConnection.conn {
				newConnection.channel.connectedUsers = append(newConnection.channel.connectedUsers[:i], newConnection.channel.connectedUsers[i+1:]...)
				break
			}
		}
		newConnection.channel.mu.Unlock()

		logger.Info(fmt.Sprintf("User %s left channel: %s", newConnection.user, newConnection.channel.name))
		LogChannelUsers(newConnection.channel)
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
			logger.Error("Failed to read message from the WebSocket connection: ", err)
			return
		}

		// Unmarshal JSON message into the Message struct
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Error("Failed to unmarshal JSON message: ", err)
			continue
		}

		msg.Timestamp = time.Now()
		msg.UserID = newConnection.user

		// set message type to message if it is not present
		if msg.Type == "" {
			msg.Type = "message"
		}

		if msg.Type == "changeChannel" {
			logger.Info(fmt.Sprintf("Received changeChannel request from user %s: %s", msg.UserID, msg.Content))
		}

		// Broadcast message to all connected users in the channel
		currentChannel := newConnection.channel

		go func() {
			currentChannel.mu.Lock()
			logger.Info(fmt.Sprintf("Sending message to channel: %s | "+msg.Timestamp.Format(time.RFC3339)+" "+msg.UserID+": "+msg.Content, currentChannel.name))
			for _, connection := range currentChannel.connectedUsers {
				err = connection.conn.WriteMessage(websocket.TextMessage, []byte(msg.Timestamp.Format(time.RFC3339)+" "+msg.UserID+": "+msg.Content))
				if err != nil {
					logger.Error("Failed to write message to the WebSocket connection: ", err)
					return
				}
			}
			currentChannel.mu.Unlock()
		}()
	}
}
