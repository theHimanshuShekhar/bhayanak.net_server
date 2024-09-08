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

// Add a read-write mutex for channel safety
var channelsMu sync.RWMutex

// Upgrader specifies the parameters for upgrading an HTTP connection to a WebSocket connection.
var upgrader = websocket.Upgrader{}

func joinChannel(conn *Connection, channelName string) {
	// Lock channels for reading
	channelsMu.RLock()
	var targetChannel *Channel
	for i := range channels {
		if channels[i].name == channelName {
			targetChannel = &channels[i]
			break
		}
	}
	channelsMu.RUnlock()

	if targetChannel == nil {
		// Channel not found
		logger.Error(fmt.Sprintf("Channel %s not found", channelName), nil)
		return
	}

	// Lock the current channel before removing the connection
	if conn.channel != nil {
		conn.channel.mu.Lock()
		// Remove the connection from the current channel
		for i, connection := range conn.channel.connectedUsers {
			if connection.conn == conn.conn {
				conn.channel.connectedUsers = append(conn.channel.connectedUsers[:i], conn.channel.connectedUsers[i+1:]...)
				break
			}
		}
		conn.channel.mu.Unlock()
	}

	// keep reference to the current channel
	previousChannel := conn.channel

	// Update the connection's channel reference
	conn.channel = targetChannel

	// Lock the new channel before adding the connection
	conn.channel.mu.Lock()
	conn.channel.connectedUsers = append(conn.channel.connectedUsers, conn)
	conn.channel.mu.Unlock()

	logger.Info(fmt.Sprintf("User %s left channel: %s", conn.user, previousChannel.name))
	logger.Info(fmt.Sprintf("User %s joined channel: %s", conn.user, conn.channel.name))

	LogChannelUsers(previousChannel)
	LogChannelUsers(conn.channel)
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
	currentConnection := Connection{
		conn: conn,
		user: uuid.NewString(),
	}

	// By default, join the general channel
	joinChannel(&currentConnection, "general")

	// Defer the removal of the connection after the user disconnects
	defer func() {
		// Remove the connection from the channel safely
		currentConnection.channel.mu.Lock()
		for i, connection := range currentConnection.channel.connectedUsers {
			if connection.conn == currentConnection.conn {
				currentConnection.channel.connectedUsers = append(currentConnection.channel.connectedUsers[:i], currentConnection.channel.connectedUsers[i+1:]...)
				break
			}
		}
		currentConnection.channel.mu.Unlock()

		logger.Info(fmt.Sprintf("User %s left channel: %s", currentConnection.user, currentConnection.channel.name))
		LogChannelUsers(currentConnection.channel)
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
		msg.UserID = currentConnection.user

		// set message type to message if it is not present
		if msg.Type == "" {
			msg.Type = "message"
		}

		if msg.Type == "changeChannel" {
			logger.Info(fmt.Sprintf("Received changeChannel request from user %s: %s", msg.UserID, msg.Content))

			joinChannel(&currentConnection, msg.Content)
		}

		if msg.Type == "message" {
			// Broadcast message to all connected users in the channel
			currentChannel := currentConnection.channel

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
}

func LogChannelUsers(channel *Channel) {
	if channel == nil {
		return
	}
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
