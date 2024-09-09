package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/theHimanshuShekhar/bchat-server/pkg/logger"
	"github.com/theHimanshuShekhar/bchat-server/pkg/models"
	"github.com/theHimanshuShekhar/bchat-server/pkg/utils"
)

// List of all the channels
var channels []models.ChatRoom = []models.ChatRoom{
	{
		Name:           "general",
		ConnectedUsers: []*models.Connection{},
	},
	{
		Name:           "tech",
		ConnectedUsers: []*models.Connection{},
	},
	{
		Name:           "fun",
		ConnectedUsers: []*models.Connection{},
	},
}

// Add a read-write mutex for channel safety
var channelsMu sync.RWMutex

func JoinChatRoom(conn *models.Connection, channelName string) {
	// Lock channels for reading
	channelsMu.RLock()
	var targetChannel *models.ChatRoom
	for i := range channels {
		if channels[i].Name == channelName {
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

	// Leave the current channel if it exists
	if conn.ChatRoom != nil {
		LeaveChatRoom(conn)
	}

	// Update the connection's channel reference
	conn.ChatRoom = targetChannel

	// Lock the new channel before adding the connection
	conn.ChatRoom.Mu.Lock()
	conn.ChatRoom.ConnectedUsers = append(conn.ChatRoom.ConnectedUsers, conn)
	conn.ChatRoom.Mu.Unlock()

	logger.Info(fmt.Sprintf("User %s joined channel: %s", conn.User, conn.ChatRoom.Name))

	utils.LogChannelUsers(conn.ChatRoom)
}

func LeaveChatRoom(conn *models.Connection) {
	// Lock the current channel before removing the connection
	if conn.ChatRoom != nil {
		conn.ChatRoom.Mu.Lock()

		// keep reference to the current channel
		previousChannel := conn.ChatRoom

		// Remove the connection from the current channel
		for i, connection := range conn.ChatRoom.ConnectedUsers {
			if connection.Conn == conn.Conn {
				conn.ChatRoom.ConnectedUsers = append(conn.ChatRoom.ConnectedUsers[:i], conn.ChatRoom.ConnectedUsers[i+1:]...)
				break
			}
		}

		logger.Info(fmt.Sprintf("User %s left channel: %s", conn.User, previousChannel.Name))
		conn.ChatRoom.Mu.Unlock()
		conn.ChatRoom = nil

		// Log the updated users in the channel
		utils.LogChannelUsers(previousChannel)
	}
}

func BroadcastMessage(currentConnection *models.Connection, msg models.Message) {
	msg.Timestamp = time.Now()
	msg.UserID = currentConnection.User

	// set message type to message if it is not present
	if msg.Type == "" {
		msg.Type = "message"
	}

	if msg.Type == "changeChannel" {
		logger.Info(fmt.Sprintf("Received changeChannel request from user %s: %s", msg.UserID, msg.Content))
		JoinChatRoom(currentConnection, msg.Content)
	}

	if msg.Type == "message" {
		// Broadcast message to all connected users in the channel
		currentChannel := currentConnection.ChatRoom

		go func() {
			currentChannel.Mu.Lock()
			logger.Info(fmt.Sprintf("Sending message to channel: %s | "+msg.Timestamp.Format(time.RFC3339)+" "+msg.UserID+": "+msg.Content, currentChannel.Name))
			for _, connection := range currentChannel.ConnectedUsers {
				err := connection.Conn.WriteMessage(websocket.TextMessage, []byte(msg.Timestamp.Format(time.RFC3339)+" "+msg.UserID+": "+msg.Content))
				if err != nil {
					logger.Error("Failed to write message to the WebSocket connection: ", err)
					return
				}
			}
			currentChannel.Mu.Unlock()
		}()
	}
}
