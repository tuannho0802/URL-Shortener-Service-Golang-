package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	clients   map[*websocket.Conn]bool
	broadcast chan bool
	mutex     sync.Mutex
}

var MainHub = Hub{
	clients:   make(map[*websocket.Conn]bool),
	broadcast: make(chan bool, 100),
}

// Upgrade HTTP to WebSocket
func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	MainHub.mutex.Lock()
	MainHub.clients[conn] = true
	MainHub.mutex.Unlock()

	// Clean after disconnect
	defer func() {
		MainHub.mutex.Lock()
		delete(MainHub.clients, conn)
		MainHub.mutex.Unlock()
		conn.Close()
	}()

	// keep connect and detect client out
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *Hub) Run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	hasUpdate := false

	for {
		select {
		case <-h.broadcast:
			hasUpdate = true

		case <-ticker.C:
			if hasUpdate {
				h.mutex.Lock()
				for client := range h.clients {
					err := client.WriteJSON(gin.H{"action": "update_links"})
					if err != nil {
						client.Close()
						delete(h.clients, client)
					}
				}
				h.mutex.Unlock()
				hasUpdate = false
			}
		}
	}
}

// Notify data change
var lastNotify time.Time
var notifyMutex sync.Mutex

func NotifyDataChange() {
	notifyMutex.Lock()
	defer notifyMutex.Unlock()

	// Only allow one notify per 2s
	if time.Since(lastNotify) < 2000*time.Millisecond {
		return
	}

	lastNotify = time.Now()
	{
		select {
		case MainHub.broadcast <- true:
		default:

		}
	}
}
