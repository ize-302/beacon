// Package ws
package ws

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
)

const (
	// sendBuffer is how many frames may be queued for a client before it is
	// considered too far behind. This counts frames, not positions: at one frame
	// per write this is many seconds of backlog, and it bounds a slow client's
	// memory at buffer * batch size rather than letting a single large batch
	// overrun a per-point queue.
	sendBuffer = 64

	// writeWait bounds a single socket write, so a stalled client cannot park
	// its writer goroutine indefinitely.
	writeWait = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// client owns one connection. Only its writer goroutine ever writes to conn
type client struct {
	conn *websocket.Conn
	send chan gpspoints.PositionFrame

	// done is closed exactly once to signal the writer goroutine to stop. The
	// send channel is deliberately never closed: Broadcast may be mid-send from
	// another goroutine, and closing underneath it would panic.
	done      chan struct{}
	closeOnce sync.Once
}

func (c *client) close() {
	c.closeOnce.Do(func() {
		close(c.done)
		c.conn.Close()
	})
}

// Hub tracks connected clients. Its mutex guards the map and nothing else — it
// is never held across a network write.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*client]struct{}),
	}
}

func (h *Hub) add(conn *websocket.Conn) *client {
	c := &client{
		conn: conn,
		send: make(chan gpspoints.PositionFrame, sendBuffer),
		done: make(chan struct{}),
	}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	go h.writeLoop(c)
	return c
}

func (h *Hub) remove(c *client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()

	c.close()
}

func (h *Hub) count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) writeLoop(c *client) {
	defer h.remove(c)

	for {
		select {
		case m := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.conn.WriteJSON(m); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

// Broadcast hands the frame to every client's queue and returns. It performs no
// I/O, so one slow client cannot delay any other.
func (h *Hub) Broadcast(m gpspoints.PositionFrame) {
	h.mu.RLock()
	clients := make([]*client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		select {
		case c.send <- m:
		case <-c.done:
			// client is going away; nothing to do
		default:
			// queue full: drop this frame rather than stall the broadcast.
			// The next one supersedes it anyway.
		}
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

	c := h.Hub.add(conn)
	defer h.Hub.remove(c)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
