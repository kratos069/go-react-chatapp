package utils

import (
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// Map to store online users: {userId: websocket.Conn}
var userSocketMap sync.Map // sync.Map for thread-safe access

// GetReceiverSocket retrieves the WebSocket connection for a given user ID
func GetReceiverSocket(userId int) *websocket.Conn {
	value, ok := userSocketMap.Load(userId)
	if ok {
		return value.(*websocket.Conn)
	}
	return nil
}

// WebSocketHandler establishes a WebSocket connection and manages events
func WebSocketHandler(conn *websocket.Conn) {
	defer conn.Close()

	token := conn.Cookies(CookieName)
	if token == "" {
		log.Println("Authentication token missing")
		return
	}

	// Validate the token (assuming ValidateToken is implemented)
	claims, err := ValidateToken(token)
	if err != nil {
		log.Printf("Invalid token: %v", err)
		return
	}

	// Extract user ID from token claims
	userIdFloat, ok := claims["id"].(float64)
	if !ok {
		log.Println("Invalid user ID in token claims")
		return
	}
	userId := int(userIdFloat)

	// Store the connection in the userSocketMap
	userSocketMap.Store(userId, conn)
	log.Printf("User connected: %s, UserID: %d\n", conn.RemoteAddr(),
		userId)

	// Broadcast updated list of online users
	broadcastOnlineUsers()

	// Keep reading messages from the WebSocket
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error for user %d: %v\n", userId, err)
			break
		}
	}

	// Remove the user from the map when disconnected
	userSocketMap.Delete(userId)
	log.Printf("User disconnected: %s, UserID: %d\n", conn.RemoteAddr(), userId)

	// Broadcast updated list of online users
	broadcastOnlineUsers()
}

// broadcastOnlineUsers broadcasts the list of online users to all connected clients
func broadcastOnlineUsers() {
	onlineUsers := make([]int, 0)

	// Collect all user IDs
	userSocketMap.Range(func(key, value interface{}) bool {
		userId := key.(int)
		onlineUsers = append(onlineUsers, userId)
		return true
	})

	// Use goroutines for broadcasting to avoid blocking
	userSocketMap.Range(func(key, value interface{}) bool {
		go func(conn *websocket.Conn) {
			err := conn.WriteJSON(map[string]interface{}{
				"event":       "getOnlineUsers",
				"onlineUsers": onlineUsers,
			})
			if err != nil {
				log.Printf("Error broadcasting online users to user %d: %v\n", key, err)
				userSocketMap.Delete(key) // Remove stale connection
			}
		}(value.(*websocket.Conn))
		return true
	})
}

// RemoveReceiverSocket removes a WebSocket connection for a given user ID
func RemoveReceiverSocket(userId int) {
	userSocketMap.Delete(userId)
}
