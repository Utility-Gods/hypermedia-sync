# Hypermedia-Driven Real-Time Sync with SSE + HTMX

## The Big Idea

Instead of sending JSON data and letting clients figure out how to display it, send complete HTML from the server. Real-time updates become simple DOM replacements.

**Traditional:** Action → JSON → Client Processing → DOM Updates  
**Hypermedia:** Action → HTML → Direct DOM Replacement

## How It Works

1. **User Action**: Click button in Browser A
2. **Server Processing**: Handle business logic, generate fresh HTML
3. **Dual Response**:
   - Send updated HTML to Browser A (HTMX response)
   - Broadcast same HTML to all other browsers (SSE event)
4. **Result**: All browsers show identical UI instantly

## Core Pattern

```html
<!-- SSE Connection Wrapper (never gets replaced) -->
<div hx-ext="sse" 
     sse-connect="/events" 
     sse-swap="content-updated"
     hx-target="#content">
  <div id="content">
    <!-- This content gets swapped -->
  </div>
</div>
```

## Server Implementation

```go
// SSE Hub manages connections
type Hub struct {
    connections map[string]*Connection
    broadcast   chan Event
    register    chan *Connection
    unregister  chan *Connection
}

// Handle user action
func actionHandler(c echo.Context) error {
    // 1. Process business logic
    updateData()
    
    // 2. Generate HTML
    html := renderUpdatedContent()
    
    // 3. Broadcast to SSE connections
    hub.broadcast <- Event{
        Name: "content-updated",
        Data: html,
    }
    
    // 4. Return HTML to originator
    return c.HTML(200, html)
}
```

## Critical Implementation Details

### SSE Data Formatting
**The most common mistake:** SSE requires single-line data or proper multiline formatting.

```go
// ❌ Wrong: Multiline HTML breaks SSE parsing
html := `<div>
    <p>Content</p>
</div>`

// ✅ Correct: Single line or properly formatted
html := `<div><p>Content</p></div>`

// Or handle multiline properly:
eventData := strings.ReplaceAll(html, "\n", "\ndata: ")
fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, eventData)
```

### Originator Filtering
Prevent echo effects by excluding the action originator from SSE broadcasts.

```javascript
// Generate unique ID per browser tab
window.originatorId = crypto.randomUUID();

// Add to all HTMX requests
document.addEventListener('htmx:configRequest', function(evt) {
    evt.detail.headers['X-Originator-ID'] = window.originatorId;
});
```

```go
// Server excludes originator from broadcast
originatorID := c.Request().Header.Get("X-Originator-ID")
hub.BroadcastExcluding(event, originatorID)
```

## Benefits

1. **Simple Mental Model**: Server controls all UI, clients just swap HTML
2. **Zero Client Logic**: No JSON parsing, no DOM manipulation code
3. **Framework Agnostic**: Works with any server that generates HTML
4. **Easy Debugging**: Inspect HTML directly, no complex state management
5. **Consistent UI**: All clients see identical representation

## Trade-offs

1. **Bandwidth**: HTML is more verbose than JSON
2. **Server Load**: Template rendering on every update
3. **Limited Client Customization**: UI logic lives on server

## When to Use

**✅ Great for:**
- Admin dashboards with live updates
- Team management interfaces
- Collaborative applications
- Live data monitoring

**❌ Consider alternatives for:**
- High-frequency updates (>10/second)
- Mobile apps with offline requirements
- Complex client-side interactions
- Games or real-time graphics

## Semantic Component Separation

**HTMX Components** (for action originators):
- Complete interaction responses
- Include toasts, modals, navigation
- Handle full user workflow

**SSE Components** (for broadcasts):
- Pure content updates
- No side effects or notifications
- Just the current state representation

```go
// Different templates for different purposes
htmxResponse := CompleteInteractionResponse(data, toasts, modals)
sseContent := PureDataRepresentation(data)
```

## Architecture Components

1. **SSE Hub**: Manages organization-scoped connections
2. **Connection Management**: Graceful registration/cleanup
3. **Event Broadcasting**: HTML payloads with semantic event names
4. **Originator Filtering**: Prevents action echo effects
5. **Dual Response System**: HTMX + SSE responses

## Quick Start

1. Set up SSE hub with connection management
2. Create HTML templates for updates
3. Add SSE wrapper elements with `hx-ext="sse"`
4. Implement dual response handlers
5. Generate unique originator IDs per browser tab

The pattern turns real-time sync from a complex state management problem into simple HTML broadcasting. The server becomes the single source of truth for how things should look, eliminating client-side UI complexity.