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
			h.connMu.Unlock()

		case conn := <-h.unregister:
			h.connMu.Lock()
			if _, exists := h.connections[conn.ID]; exists {
				delete(h.connections, conn.ID)
			}
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

	return fmt.Sprintf(`<input type="checkbox" id="cb-%d" %s hx-post="/toggle/%d" hx-swap="none"><label for="cb-%d"> %d</label>`,
		id, checkedAttr, id, id, id)
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
    <script src="/static/js/htmx.js"></script>
    <script src="/static/js/sse.js"></script>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap" rel="stylesheet">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body { 
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif; 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            color: #1a202c;
        }
        
        .hero {
            text-align: center;
            padding: 2rem 1rem 3rem;
            color: white;
        }
        
        .hero h1 {
            font-size: 3rem;
            font-weight: 700;
            margin-bottom: 0.5rem;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
        }
        
        .hero .subtitle {
            font-size: 1.25rem;
            opacity: 0.9;
            margin-bottom: 2rem;
            font-weight: 300;
        }
        
        .github-link {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0.75rem 1.5rem;
            background: white;
            color: #667eea;
            text-decoration: none;
            border-radius: 2rem;
            font-weight: 600;
            transition: transform 0.2s, box-shadow 0.2s;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        
        .github-link:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 12px rgba(0,0,0,0.15);
        }
        
        .checkbox-container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 0 1rem;
        }
        
        .checkbox-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
            gap: 0.5rem;
            max-height: 60vh;
            overflow-y: auto;
            padding: 1.5rem;
            background: white;
            border-radius: 1rem;
            box-shadow: 0 10px 25px rgba(0,0,0,0.1);
        }
        
        .checkbox-grid::-webkit-scrollbar {
            width: 12px;
        }
        
        .checkbox-grid::-webkit-scrollbar-track {
            background: #f1f1f1;
            border-radius: 10px;
        }
        
        .checkbox-grid::-webkit-scrollbar-thumb {
            background: #667eea;
            border-radius: 10px;
        }
        
        .checkbox-grid::-webkit-scrollbar-thumb:hover {
            background: #5a67d8;
        }
        
        .checkbox-item {
            display: flex;
            align-items: center;
            padding: 0.5rem;
            background: #f7fafc;
            border-radius: 0.5rem;
            transition: all 0.2s;
            border: 1px solid #e2e8f0;
        }
        
        .checkbox-item:hover {
            background: #edf2f7;
            transform: scale(1.05);
            border-color: #cbd5e0;
        }
        
        .checkbox-item input[type="checkbox"] {
            width: 18px;
            height: 18px;
            margin-right: 0.5rem;
            cursor: pointer;
            accent-color: #667eea;
        }
        
        .checkbox-item label {
            cursor: pointer;
            font-size: 0.875rem;
            color: #4a5568;
            user-select: none;
        }
        
        .footer {
            text-align: center;
            padding: 3rem 1rem;
            color: white;
            opacity: 0.8;
        }
        
        @media (max-width: 768px) {
            .hero h1 {
                font-size: 2rem;
            }
            
            .checkbox-grid {
                grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
                max-height: 50vh;
            }
        }
    </style>
</head>
<body>
    <div class="hero">
        <h1>10,000 Checkboxes</h1>
        <p class="subtitle">A Real-Time Hypermedia Experiment</p>
        
        <a href="https://github.com/Utility-Gods/hypermedia-sync" class="github-link" target="_blank">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
            </svg>
            View on GitHub
        </a>
    </div>

    <div class="checkbox-container">
        <!-- SSE Connection Wrapper -->
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
                    <label for="cb-{{.ID}}">{{.ID}}</label>
                </div>
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
	}

	var cbData []CheckboxData

	for i := 1; i <= 10000; i++ {
		checked := checkboxes[i]
		cbData = append(cbData, CheckboxData{ID: i, Checked: checked})
	}

	// Generate unique originator ID for this page load
	originatorID := fmt.Sprintf("page-%d-%d", time.Now().UnixNano(), rand.Intn(1000000))

	data := PageData{
		Checkboxes:   cbData,
		OriginatorID: originatorID,
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
