package utils

import (
	"fmt"
	"strings"

	"github.com/theHimanshuShekhar/bchat-server/pkg/logger"
	"github.com/theHimanshuShekhar/bchat-server/pkg/models"
)

func LogChannelUsers(channel *models.ChatRoom) {
	if channel == nil {
		return
	}
	logger.Info(fmt.Sprintf("Channel: %s | Connected Users: %d", channel.Name, len(channel.ConnectedUsers)))

	// Use a slice to collect the users' IDs
	var users []string
	// Lock the channel before accessing the users if necessary
	channel.Mu.Lock()
	for _, connection := range channel.ConnectedUsers {
		users = append(users, "["+connection.User+"]")
	}
	channel.Mu.Unlock()
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
