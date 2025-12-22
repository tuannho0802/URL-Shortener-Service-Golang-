package handlers

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/middleware"
)

// Config HTTP WebSocket Upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Avoid CORS issue
	},
}

// Manage Client
type Hub struct {
	// Manage clients
	clients map[uint][]*websocket.Conn
	// channel to broadcast data
	broadcast chan uint
	mutex     sync.Mutex
}

var MainHub = Hub{
	clients:   make(map[uint][]*websocket.Conn),
	broadcast: make(chan uint, 100),
}

// Upgrade HTTP to WebSocket
func HandleWebSocket(c *gin.Context) {
	// get token from query string
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
		return
	}

	// Call validateToken function
	claims, err := middleware.ValidateToken(tokenString)
	if err != nil {
		log.Printf("WS Auth Error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// get user_id
	userID := claims.UserID

	// Upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade Failed: %v", err)
		return
	}

	MainHub.mutex.Lock()
	MainHub.clients[userID] = append(MainHub.clients[userID], conn)
	MainHub.mutex.Unlock()

	defer func() {
		MainHub.mutex.Lock()
		connections := MainHub.clients[userID]
		for i, v := range connections {
			if v == conn {
				MainHub.clients[userID] = append(connections[:i], connections[i+1:]...)
				break
			}
		}
		// if not connections, delete
		if len(MainHub.clients[userID]) == 0 {
			delete(MainHub.clients, userID)
		}
		MainHub.mutex.Unlock()
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
func (h *Hub) Run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Map to group userID for update ticker
	pendingUpdates := make(map[uint]bool)

	for {
		select {
		case userID := <-h.broadcast:
			pendingUpdates[userID] = true

		case <-ticker.C:
			if len(pendingUpdates) > 0 {
				h.mutex.Lock()
				for userID := range pendingUpdates {
					if connections, ok := h.clients[userID]; ok {
						for _, client := range connections {
							err := client.WriteJSON(gin.H{"action": "update_links"})
							if err != nil {
								// If an error , close and delete will be handled in the HandleWebSocket defer
								continue
							}
						}
					}
				}
				h.mutex.Unlock()
				// reset list after update
				pendingUpdates = make(map[uint]bool)
			}
		}
	}
}

// Notify data change
var lastNotify time.Time
var notifyMutex sync.Mutex

// get userID to notify
func NotifyDataChange(userID uint) {
	notifyMutex.Lock()
	defer notifyMutex.Unlock()

	// Only allow one notify per 2s
	if time.Since(lastNotify) < 2000*time.Millisecond {
		return
	}

	lastNotify = time.Now()
	{
		select {
		case MainHub.broadcast <- userID:
		default:
		}
	}
}
