package sse

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// Hub manages SSE connections and broadcasts
type Hub struct {
	connections map[string]*Connection
	broadcast   chan Event
	register    chan *Connection
	unregister  chan *Connection
	connMu      sync.RWMutex
	onlineCount int
}

type Connection struct {
	ID     string
	Writer http.ResponseWriter
	Done   chan struct{}
}

type Event struct {
	Name      string
	Data      string
	ExcludeID string // Originator ID to exclude from broadcast
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*Connection),
		broadcast:   make(chan Event, 100),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.connMu.Lock()
			h.connections[conn.ID] = conn
			h.onlineCount = len(h.connections)
			onlineCount := h.onlineCount
			h.connMu.Unlock()
			
			// Broadcast online count update to all connections
			onlineHTML := fmt.Sprintf(`<span class="relative flex h-2 w-2">
				<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
				<span class="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
			</span>
			<span class="text-secondary-50 font-semibold text-xs sm:text-sm">
				<span class="hidden sm:inline">%d users online</span>
				<span class="sm:hidden">%d</span>
			</span>`, onlineCount, onlineCount)
			h.broadcast <- Event{
				Name: "online-count-updated",
				Data: onlineHTML,
			}

		case conn := <-h.unregister:
			h.connMu.Lock()
			if _, exists := h.connections[conn.ID]; exists {
				delete(h.connections, conn.ID)
				h.onlineCount = len(h.connections)
				onlineCount := h.onlineCount
				h.connMu.Unlock()
				
				// Broadcast online count update to all connections
				onlineHTML := fmt.Sprintf(`<span class="relative flex h-2 w-2">
					<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
					<span class="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
				</span>
				<span class="text-secondary-50 font-semibold text-xs sm:text-sm">
					<span class="hidden sm:inline">%d users online</span>
					<span class="sm:hidden">%d</span>
				</span>`, onlineCount, onlineCount)
				h.broadcast <- Event{
					Name: "online-count-updated",
					Data: onlineHTML,
				}
			} else {
				h.connMu.Unlock()
			}

		case event := <-h.broadcast:
			h.connMu.RLock()
			for connID, conn := range h.connections {
				if connID != event.ExcludeID {
					go func(c *Connection) {
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("Error broadcasting to connection: %v\n", r)
							}
						}()
						
						select {
						case <-c.Done:
							return
						default:
						}
						
						if c.Writer == nil {
							fmt.Printf("Warning: Writer is nil for connection %s\n", c.ID)
							return
						}
						
						eventData := strings.ReplaceAll(event.Data, "\n", "\ndata: ")
						_, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event.Name, eventData)
						if err != nil {
							fmt.Printf("Error writing to connection %s: %v\n", c.ID, err)
							return
						}
						
						if flusher, ok := c.Writer.(http.Flusher); ok {
							flusher.Flush()
						}
					}(conn)
				}
			}
			h.connMu.RUnlock()
		}
	}
}

func (h *Hub) GetOnlineCount() int {
	h.connMu.RLock()
	defer h.connMu.RUnlock()
	return len(h.connections)
}

func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

func (h *Hub) Unregister(conn *Connection) {
	h.unregister <- conn
}

func (h *Hub) Broadcast(event Event) {
	h.broadcast <- event
}