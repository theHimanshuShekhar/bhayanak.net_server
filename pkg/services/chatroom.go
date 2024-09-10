package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/theHimanshuShekhar/bhayanak.net/pkg/logger"
	"github.com/theHimanshuShekhar/bhayanak.net/pkg/models"
	"github.com/theHimanshuShekhar/bhayanak.net/pkg/utils"
)

// List of all the chatrooms
var chatrooms []models.ChatRoom = []models.ChatRoom{
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

// Add a read-write mutex for chatroom safety
var chatRoomsMu sync.RWMutex

func JoinChatRoom(conn *models.Connection, chatRoomName string) {
	// Lock chatrooms for reading
	chatRoomsMu.RLock()
	var targetChatRoom *models.ChatRoom
	for i := range chatrooms {
		if chatrooms[i].Name == chatRoomName {
			targetChatRoom = &chatrooms[i]
			break
		}
	}
	chatRoomsMu.RUnlock()

	if targetChatRoom == nil {
		// Chatroom not found
		logger.Error(fmt.Sprintf("Chatroom %s not found", chatRoomName), nil)
		return
	}

	// Leave the current chatroom if it exists
	if conn.ChatRoom != nil {
		LeaveChatRoom(conn)
	}

	// Update the connection's chatroom reference
	conn.ChatRoom = targetChatRoom

	// Lock the new chatroom before adding the connection
	conn.ChatRoom.Mu.Lock()
	conn.ChatRoom.ConnectedUsers = append(conn.ChatRoom.ConnectedUsers, conn)
	conn.ChatRoom.Mu.Unlock()

	logger.Info(fmt.Sprintf("User %s joined chatroom: %s", conn.User, conn.ChatRoom.Name))

	utils.LogChatRoomUsers(conn.ChatRoom)
}

func LeaveChatRoom(conn *models.Connection) {
	// Lock the current chatroom before removing the connection
	if conn.ChatRoom != nil {
		conn.ChatRoom.Mu.Lock()

		// keep reference to the current chatroom
		previousChatRoom := conn.ChatRoom

		// Remove the connection from the current chatroom
		for i, connection := range conn.ChatRoom.ConnectedUsers {
			if connection.Conn == conn.Conn {
				conn.ChatRoom.ConnectedUsers = append(conn.ChatRoom.ConnectedUsers[:i], conn.ChatRoom.ConnectedUsers[i+1:]...)
				break
			}
		}

		logger.Info(fmt.Sprintf("User %s left chatroom: %s", conn.User, previousChatRoom.Name))
		conn.ChatRoom.Mu.Unlock()
		conn.ChatRoom = nil

		// Log the updated users in the chatroom
		utils.LogChatRoomUsers(previousChatRoom)
	}
}

func BroadcastMessage(currentConnection *models.Connection, msg models.Message) {
	msg.Timestamp = time.Now()
	msg.UserID = currentConnection.User

	// set message type to message if it is not present
	if msg.Type == "" {
		msg.Type = "message"
	}

	if msg.Type == "changeChatRoom" {
		logger.Info(fmt.Sprintf("Received changeChatRoom request from user %s: %s", msg.UserID, msg.Content))
		JoinChatRoom(currentConnection, msg.Content)
	}

	if msg.Type == "message" {
		// Broadcast message to all connected users in the chatroom
		currentChatRoom := currentConnection.ChatRoom

		go func() {
			currentChatRoom.Mu.Lock()
			logger.Info(fmt.Sprintf("Sending message to chatroom: %s | "+msg.Timestamp.Format(time.RFC3339)+" "+msg.UserID+": "+msg.Content, currentChatRoom.Name))
			for _, connection := range currentChatRoom.ConnectedUsers {
				err := connection.Conn.WriteMessage(websocket.TextMessage, []byte(msg.Timestamp.Format(time.RFC3339)+" "+msg.UserID+": "+msg.Content))
				if err != nil {
					logger.Error("Failed to write message to the WebSocket connection: ", err)
					return
				}
			}
			currentChatRoom.Mu.Unlock()
		}()
	}
}
