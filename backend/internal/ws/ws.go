// Package ws
package ws

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/ize-302/beacon/backend/internal/database"
	"github.com/ize-302/beacon/backend/internal/locations"
)

type Handler struct {
	*database.Handler
}

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

func (h *Hub) Broadcast(m locations.CreateLocation) {
	h.mu.RLock()

	clients := make([]*websocket.Conn, 0, len(h.clients))
	for conn := range h.clients {
		clients = append(clients, conn)
	}

	h.mu.RUnlock()

	for _, conn := range clients {
		if err := conn.WriteJSON(m); err != nil {
			h.Remove(conn)
		}
	}
}

func (h *Handler) WsHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading:", err)
			return
		}
		defer hub.Remove(conn)
		hub.Add(conn)

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("error reading message:", err)
				break
			}
		}
	}
}
