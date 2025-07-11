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
}

type Connection struct {
	ID     string
	Writer http.ResponseWriter
	Done   chan struct{}
}

type Event struct {
	Name       string
	Data       string
	ExcludeID  string // Originator ID to exclude from broadcast
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
			h.connMu.Unlock()

		case conn := <-h.unregister:
			h.connMu.Lock()
			delete(h.connections, conn.ID)
			h.connMu.Unlock()

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
						// Format SSE event data properly - replace newlines with data: prefix
						eventData := strings.ReplaceAll(event.Data, "\n", "\ndata: ")
						fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event.Name, eventData)
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
	
	return fmt.Sprintf(`<input type="checkbox" id="cb-%d" %s hx-post="/toggle/%d" hx-swap="none"><label for="cb-%d">Checkbox %d</label>`, 
		id, checkedAttr, id, id, id)
}

func main() {
	// Initialize checkboxes
	for i := 1; i <= 10; i++ {
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
    <title>10,000 Checkboxes - Hypermedia Sync Demo</title>
    <script src="/static/js/htmx.js"></script>
    <script src="/static/js/sse.js"></script>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
            padding: 20px;
            background-color: #007bff;
            color: white;
            border-radius: 8px;
        }
        .stats {
            text-align: center;
            margin-bottom: 20px;
            padding: 15px;
            background-color: #e9ecef;
            border-radius: 8px;
        }
        .checkbox-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
            gap: 5px;
            max-height: 600px;
            overflow-y: auto;
            border: 1px solid #ddd;
            padding: 20px;
            border-radius: 8px;
            background-color: #fafafa;
        }
        .checkbox-item {
            display: flex;
            align-items: center;
            padding: 8px;
            background-color: white;
            border-radius: 4px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            transition: background-color 0.2s;
        }
        .checkbox-item:hover {
            background-color: #f8f9fa;
        }
        .checkbox-item input {
            margin-right: 8px;
            transform: scale(1.2);
        }
        .checkbox-item label {
            cursor: pointer;
            user-select: none;
            font-size: 14px;
        }
        .instructions {
            margin-top: 20px;
            padding: 15px;
            background-color: #d4edda;
            border: 1px solid #c3e6cb;
            border-radius: 8px;
            color: #155724;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎯 10 Checkboxes Demo</h1>
            <p>Hypermedia-Driven Real-Time Synchronization with SSE + HTMX</p>
        </div>
        
        <div class="stats">
            <strong>Checked: <span id="checked-count">{{.CheckedCount}}</span> / 10</strong>
        </div>

        <!-- SSE Connection Wrapper - This element stays, only inner content gets replaced -->
        <div hx-ext="sse" 
             sse-connect="/events?originator={{.OriginatorID}}" 
             id="sse-wrapper">
            <div class="checkbox-grid" id="team-section">
                {{range .Checkboxes}}
                <div class="checkbox-item" id="checkbox-{{.ID}}" 
                     sse-swap="checkbox-{{.ID}}-updated"
                     hx-swap="innerHTML">
                    <input type="checkbox" 
                           id="cb-{{.ID}}" 
                           {{if .Checked}}checked{{end}}
                           hx-post="/toggle/{{.ID}}"
                           hx-swap="none">
                    <label for="cb-{{.ID}}">Checkbox {{.ID}}</label>
                </div>
                {{end}}
            </div>
        </div>

        <div class="instructions">
            <h3>📋 Instructions:</h3>
            <ul>
                <li>Open this page in multiple browser tabs or windows</li>
                <li>Click any checkbox in one tab</li>
                <li>Watch it update instantly in all other tabs via SSE!</li>
                <li>The server broadcasts HTML updates to all connected clients</li>
            </ul>
        </div>
    </div>

    <script>
        // Server-generated originator ID
        window.originatorId = '{{.OriginatorID}}';
        
        // Add originator ID to all HTMX requests
        document.addEventListener('htmx:configRequest', function(evt) {
            evt.detail.headers['X-Originator-ID'] = window.originatorId;
        });

        // Update checked count on page updates
        document.addEventListener('htmx:afterSwap', function(evt) {
            updateCheckedCount();
        });

        function updateCheckedCount() {
            const checked = document.querySelectorAll('input[type="checkbox"]:checked').length;
            const countElement = document.getElementById('checked-count');
            if (countElement) {
                countElement.textContent = checked;
            }
        }

        // Initial count
        updateCheckedCount();
    </script>
</body>
</html>`

	mu.RLock()
	defer mu.RUnlock()

	type PageData struct {
		Checkboxes   []CheckboxData
		CheckedCount int
		OriginatorID string
	}

	var cbData []CheckboxData
	checkedCount := 0

	for i := 1; i <= 10; i++ {
		checked := checkboxes[i]
		cbData = append(cbData, CheckboxData{ID: i, Checked: checked})
		if checked {
			checkedCount++
		}
	}

	// Generate unique originator ID for this page load
	originatorID := fmt.Sprintf("page-%d-%d", time.Now().UnixNano(), rand.Intn(1000000))
	
	data := PageData{
		Checkboxes:   cbData,
		CheckedCount: checkedCount,
		OriginatorID: originatorID,
	}

	t := template.Must(template.New("index").Parse(tmpl))
	return t.Execute(c.Response(), data)
}

func sseHandler(c echo.Context) error {
	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get originator ID from query params
	originatorID := c.QueryParam("originator")
	if originatorID == "" {
		// Fallback if no originator ID provided
		hub.connMu.RLock()
		connCount := len(hub.connections)
		hub.connMu.RUnlock()
		originatorID = fmt.Sprintf("sse-%d", connCount)
	}

	conn := &Connection{
		ID:     originatorID,
		Writer: w,
		Done:   make(chan struct{}),
	}

	// Flush headers immediately to establish connection
	if flusher, ok := w.Writer.(http.Flusher); ok {
		flusher.Flush()
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

	if id < 1 || id > 10 {
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
	// Broadcast only the affected checkbox
	hub.broadcast <- Event{
		Name:      fmt.Sprintf("checkbox-%d-updated", id),
		Data:      checkboxHTML,
		ExcludeID: originatorID,
	}

	// Return no content since we're using hx-swap="none"
	return c.NoContent(204)
}
