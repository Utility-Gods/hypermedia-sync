package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Global state for checkboxes
var (
	checkboxes = make(map[int]bool)
	mu         sync.RWMutex
)

// CheckboxData represents a checkbox state
type CheckboxData struct {
	ID      int
	Checked bool
}

// SSE Hub for managing connections
type Hub struct {
	connections map[string]*Connection
	broadcast   chan Event
	register    chan *Connection
	unregister  chan *Connection
	connMu      sync.RWMutex
	onlineCount int // Track the number of online users
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
			onlineHTML := fmt.Sprintf(`<span class="online-count">%d users online</span>`, onlineCount)
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
				onlineHTML := fmt.Sprintf(`<span class="online-count">%d users online</span>`, onlineCount)
				h.broadcast <- Event{
					Name: "online-count-updated",
					Data: onlineHTML,
				}
			} else {
				h.connMu.Unlock()
			}

		case event := <-h.broadcast:
			h.connMu.RLock()
			// Broadcast to all connections except the excluded ID
			for connID, conn := range h.connections {
				if connID != event.ExcludeID {
					go func(c *Connection) {
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("Error broadcasting to connection: %v\n", r)
							}
						}()
						
						// Check if connection is still valid
						select {
						case <-c.Done:
							// Connection is closed, skip
							return
						default:
							// Connection is active, proceed with broadcast
						}
						
						// Check if writer is not nil
						if c.Writer == nil {
							fmt.Printf("Warning: Writer is nil for connection %s\n", c.ID)
							return
						}
						
						// Format SSE event data properly - replace newlines with data: prefix
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

var hub = NewHub()

// generateSingleCheckboxHTML creates HTML for a single checkbox
func generateSingleCheckboxHTML(id int, checked bool) string {
	checkedAttr := ""
	if checked {
		checkedAttr = "checked"
	}

	return fmt.Sprintf(`<input type="checkbox" id="cb-%d" %s hx-post="/toggle/%d" hx-swap="none"><span>%d</span>`,
		id, checkedAttr, id, id)
}

func main() {
	// Initialize checkboxes
	for i := 1; i <= 10000; i++ {
		checkboxes[i] = false
	}

	go hub.Run()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Serve static files
	e.Static("/static", "static")

	// Routes
	e.GET("/", indexHandler)
	e.GET("/health", healthHandler)
	e.GET("/events", sseHandler)
	e.POST("/toggle/:id", toggleHandler)

	// Start server on port from environment or 8080
	port := ":8080"
	fmt.Printf("Server starting on %s\n", port)
	e.Logger.Fatal(e.Start(port))
}

func indexHandler(c echo.Context) error {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>One Million Checkboxes - Hypermedia Sync Experiment</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="static/js/htmx.js"></script>
    <script src="static/js/sse.js"></script>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="static/css/styles.css">
</head>
<body>
    <!-- SSE Connection Wrapper -->
    <div hx-ext="sse" 
         sse-connect="/events?originator={{.OriginatorID}}" 
         id="sse-wrapper">
         
        <div class="hero">
            <h1>10,000 Checkboxes</h1>
            <p class="subtitle">A Real-Time Hypermedia Experiment</p>
            
            <a href="https://github.com/Utility-Gods/hypermedia-sync" class="github-link" target="_blank">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                </svg>
                View on GitHub
            </a>
            
            <!-- Online Users Display -->
            <div class="online-users" id="online-users-container" sse-swap="online-count-updated" hx-swap="innerHTML">
                <span class="online-count">{{.OnlineCount}} users online</span>
            </div>
        </div>

        <div class="checkbox-container">
            <div class="checkbox-grid" id="team-section">
                {{range .Checkboxes}}
                <label for="cb-{{.ID}}" class="checkbox-item" id="checkbox-{{.ID}}"
                     sse-swap="checkbox-{{.ID}}-updated"
                     hx-swap="innerHTML">
                    <input type="checkbox" 
                           id="cb-{{.ID}}" 
                           {{if .Checked}}checked{{end}}
                           hx-post="/toggle/{{.ID}}"
                           hx-swap="none">
                    <span>{{.ID}}</span>
                </label>
                {{end}}
            </div>
        </div>
    </div>
    
    <div class="footer">
        <p>Built with HTMX + Server-Sent Events • No WebSockets, No JSON, Just HTML</p>
    </div>

    <script>
        // Server-generated originator ID
        window.originatorId = '{{.OriginatorID}}';
        
        // Add originator ID to all HTMX requests
        document.addEventListener('htmx:configRequest', function(evt) {
            evt.detail.headers['X-Originator-ID'] = window.originatorId;
        });
    </script>
</body>
</html>`

	mu.RLock()
	defer mu.RUnlock()

	type PageData struct {
		Checkboxes   []CheckboxData
		OriginatorID string
		OnlineCount  int
	}

	var cbData []CheckboxData

	for i := 1; i <= 10000; i++ {
		checked := checkboxes[i]
		cbData = append(cbData, CheckboxData{ID: i, Checked: checked})
	}

	// Generate unique originator ID for this page load
	originatorID := fmt.Sprintf("page-%d-%d", time.Now().UnixNano(), rand.Intn(1000000))

	// Get current online count
	hub.connMu.RLock()
	onlineCount := len(hub.connections)
	hub.connMu.RUnlock()

	data := PageData{
		Checkboxes:   cbData,
		OriginatorID: originatorID,
		OnlineCount:  onlineCount,
	}

	t := template.Must(template.New("index").Parse(tmpl))
	return t.Execute(c.Response(), data)
}

func sseHandler(c echo.Context) error {
	// Get originator ID from query params
	originatorID := c.QueryParam("originator")
	if originatorID == "" {
		// Fallback if no originator ID provided
		hub.connMu.RLock()
		connCount := len(hub.connections)
		hub.connMu.RUnlock()
		originatorID = fmt.Sprintf("sse-%d", connCount)
	}

	// Set headers before writing response
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("X-Accel-Buffering", "no") // Disable proxy buffering

	// Write status and flush headers
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Flush()

	// Send initial connection message
	fmt.Fprintf(c.Response().Writer, ": connected\n\n")
	c.Response().Flush()

	conn := &Connection{
		ID:     originatorID,
		Writer: c.Response().Writer,
		Done:   make(chan struct{}),
	}

	hub.register <- conn
	defer func() {
		hub.unregister <- conn
		close(conn.Done)
	}()

	// Keep connection alive
	<-c.Request().Context().Done()
	return nil
}

func toggleHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.String(400, "Invalid checkbox ID")
	}

	if id < 1 || id > 10000 {
		return c.String(400, "Checkbox ID out of range")
	}

	// Get originator ID
	originatorID := c.Request().Header.Get("X-Originator-ID")

	// Toggle checkbox state
	mu.Lock()
	checkboxes[id] = !checkboxes[id]
	newState := checkboxes[id]
	mu.Unlock()

	// Generate HTML for just this checkbox
	checkboxHTML := generateSingleCheckboxHTML(id, newState)
	// Broadcast only the affected checkbox (excluding originator)
	hub.broadcast <- Event{
		Name:      fmt.Sprintf("checkbox-%d-updated", id),
		Data:      checkboxHTML,
		ExcludeID: originatorID,
	}

	// Return no content since we're using hx-swap="none"
	return c.NoContent(204)
}

func healthHandler(c echo.Context) error {
	return c.JSON(200, map[string]string{
		"status": "healthy",
		"service": "hypermedia-sync",
	})
}
