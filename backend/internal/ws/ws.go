// Package ws
package ws

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (h *Hub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = struct{}{}
}

func (h *Hub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	conn.Close()
}

func (h *Hub) Broadcast(m gpspoints.CreateGpsPoint) {
	h.mu.RLock()
	clients := make([]*websocket.Conn, 0, len(h.clients))
	for conn := range h.clients {
		clients = append(clients, conn)
	}
	h.mu.RUnlock()

	for _, conn := range clients {
		h.mu.Lock()
		err := conn.WriteJSON(m)
		if err != nil {
			h.Remove(conn)
		}
		h.mu.Unlock()
	}
}

type WsHandler struct {
	Hub *Hub
}

func NewWsHandler(hub *Hub) *WsHandler {
	return &WsHandler{Hub: hub}
}

func (h *WsHandler) RegisterRoutes(router chi.Router) {
	router.Get("/ws", h.handleConnection)
}

func (h *WsHandler) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer h.Hub.Remove(conn)
	h.Hub.Add(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("error reading message:", err)
			break
		}
	}
}
