// Package ws pushes live check results to connected dashboard clients.
package ws

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

type Hub struct {
	mu       sync.Mutex
	conns    map[*websocket.Conn]struct{}
	snapshot func() []byte
}

// NewHub builds a hub. snapshot returns the initial payload sent on connect.
func NewHub(snapshot func() []byte) *Hub {
	return &Hub{conns: make(map[*websocket.Conn]struct{}), snapshot: snapshot}
}

// Broadcast writes b to every connected client, dropping any that error.
func (h *Hub) Broadcast(b []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.conns {
		_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
			c.Close()
			delete(h.conns, c)
		}
	}
}

// Handler upgrades the request, sends the snapshot, and holds the connection
// open until the client disconnects.
func (h *Hub) Handler() echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		if h.snapshot != nil {
			_ = conn.WriteMessage(websocket.TextMessage, h.snapshot())
		}
		h.mu.Lock()
		h.conns[conn] = struct{}{}
		h.mu.Unlock()

		// Read (and discard) until the client goes away, then clean up.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
		h.mu.Lock()
		delete(h.conns, conn)
		h.mu.Unlock()
		conn.Close()
		return nil
	}
}
