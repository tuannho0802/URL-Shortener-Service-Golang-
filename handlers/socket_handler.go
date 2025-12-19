package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Config HTTP WebSocket Upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow CORS
	},
}

// Manage Client
type Hub struct {
	clients   map[*websocket.Conn]bool
	broadcast chan bool
	mutex     sync.Mutex
}

var MainHub = Hub{
	clients:   make(map[*websocket.Conn]bool),
	broadcast: make(chan bool),
}

// Func call Hub to listen broadcast
func (h *Hub) Run() {
	for {
		<-h.broadcast
		h.mutex.Lock()
		for client := range h.clients {
			err := client.WriteJSON(gin.H{"action": "update_links"})
			if err != nil {
				client.Close()
				delete(h.clients, client)
			}
		}
		h.mutex.Unlock()
	}
}

// Handle WebSocket
func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	MainHub.mutex.Lock()
	MainHub.clients[conn] = true
	MainHub.mutex.Unlock()

	// Keep connection open
	defer func() {
		MainHub.mutex.Lock()
		delete(MainHub.clients, conn)
		MainHub.mutex.Unlock()
		conn.Close()
	}()

	// Listen message
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// Func notify data change
func NotifyDataChange() {
	MainHub.broadcast <- true
}
