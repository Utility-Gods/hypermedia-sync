# Understanding Hypermedia-Driven Real-Time Sync with SSE + HTMX

## The Big Idea

Instead of sending JSON data and letting clients figure out how to display it, send complete HTML from the server. Real-time updates become simple DOM replacements.

**Traditional:** Action → JSON → Client Processing → DOM Updates  
**Hypermedia:** Action → HTML → Direct DOM Replacement

## How It Works

1. **User Action**: Click checkbox in Browser A
2. **Server Processing**: Handle business logic, generate fresh HTML
3. **Response Strategy**:
   - Return 204 No Content to Browser A (with `hx-swap="none"`)
   - Broadcast HTML to all other browsers (SSE event)
4. **Result**: All browsers show identical UI instantly

## Core Pattern

```html
<!-- SSE Connection Wrapper (never gets replaced) -->
<div hx-ext="sse" sse-connect="/events?originator={{.OriginatorID}}">
  <!-- Each element listens for its specific update event -->
  <div id="item-1" sse-swap="item-1-updated" hx-swap="innerHTML">
    <input type="checkbox" hx-post="/toggle/1" hx-swap="none">
  </div>
  <div id="item-2" sse-swap="item-2-updated" hx-swap="innerHTML">
    <input type="checkbox" hx-post="/toggle/2" hx-swap="none">
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
func toggleHandler(c echo.Context) error {
    // 1. Get originator ID and item ID
    originatorID := c.Request().Header.Get("X-Originator-ID")
    itemID := c.Param("id")
    
    // 2. Update state
    newState := toggleItem(itemID)
    
    // 3. Generate HTML for just this item
    html := renderSingleItem(itemID, newState)
    
    // 4. Broadcast only the affected item
    hub.broadcast <- Event{
        Name:      fmt.Sprintf("item-%s-updated", itemID),
        Data:      html,
        ExcludeID: originatorID,
    }
    
    // 5. Return no content (using hx-swap="none")
    return c.NoContent(204)
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

The simplest approach to prevent duplicate updates is to generate unique originator IDs on the server and use them directly as SSE connection identifiers.

#### How Originator Filtering Works

##### 1. Server Generates Unique ID
When serving the initial page, the server generates a unique ID for that page load:

```go
// Generate unique originator ID for this page load
originatorID := fmt.Sprintf("page-%d-%d", time.Now().UnixNano(), rand.Intn(1000000))
```

##### 2. ID Embedded in HTML
The ID is passed to the template and embedded directly in the HTML:

```html
<!-- SSE connection uses server-generated ID -->
<div hx-ext="sse" 
     sse-connect="/events?originator={{.OriginatorID}}" 
     sse-swap="team-updated">

<!-- JavaScript receives the ID for HTMX requests -->
<script>
    window.originatorId = '{{.OriginatorID}}';
</script>
```

##### 3. SSE Connection Uses ID
The SSE handler uses the originator ID as the connection ID:

```go
func sseHandler(c echo.Context) error {
    originatorID := c.QueryParam("originator")
    
    conn := &Connection{
        ID:     originatorID,
        Writer: w,
    }
    
    hub.register <- conn
}
```

##### 4. Broadcasting Excludes Originator
When broadcasting, the hub excludes the connection with matching ID:

```go
type Event struct {
    Name      string
    Data      string
    ExcludeID string  // ID to exclude from broadcast
}

// In Hub.Run()
for connID, conn := range h.connections {
    if connID != event.ExcludeID {
        // Send event to this connection
    }
}
```

#### Complete Flow

1. **Page Load**: Server generates unique ID, embeds in HTML
2. **SSE Connect**: Browser connects with `?originator={id}`
3. **User Action**: HTMX sends request with `X-Originator-ID` header
4. **Server Response**: 
   - Returns 204 No Content to originator (with `hx-swap="none"`)
   - Broadcasts HTML to all SSE connections except originator
5. **Result**: Clean updates without duplicates or unnecessary data transfer

#### Benefits of Originator Filtering

- **Simplicity**: Direct ID matching, no complex topic systems
- **Server Control**: IDs generated server-side, no client coordination
- **Clean HTML**: SSE attributes work directly without JavaScript setup
- **Predictable**: Each page load gets one unique ID for its lifetime

#### Code Example

```go
// Handle checkbox toggle
func toggleHandler(c echo.Context) error {
    // Get originator ID from request
    originatorID := c.Request().Header.Get("X-Originator-ID")
    id := c.Param("id")
    
    // Update state
    newState := toggleCheckbox(id)
    
    // Generate HTML for just this checkbox
    html := fmt.Sprintf(`<input type="checkbox" id="cb-%s" %s ...>`, 
        id, checkedAttr(newState))
    
    // Broadcast only the affected element
    hub.broadcast <- Event{
        Name:      fmt.Sprintf("checkbox-%s-updated", id),
        Data:      html,
        ExcludeID: originatorID,
    }
    
    // Return no content (hx-swap="none")
    return c.NoContent(204)
}
```

## Targeted Updates Pattern

Instead of replacing entire sections, broadcast minimal HTML for specific elements:

**Benefits**:
- **Efficient**: Only affected elements are sent (50 bytes vs 2KB)
- **Performant**: Browser updates single DOM node, not entire sections
- **Scalable**: Works with thousands of items without performance degradation

**Implementation**:
```html
<!-- Each item has its own SSE event listener -->
<div id="checkbox-1" sse-swap="checkbox-1-updated" hx-swap="innerHTML">
  <input type="checkbox" ...>
</div>
```

```go
// Broadcast only what changed
hub.broadcast <- Event{
    Name: "checkbox-1-updated",
    Data: "<input type=\"checkbox\" checked ...>",
}
```

This approach:
- Sends only the changed element (~50 bytes)
- Updates only the specific DOM node
- Scales to thousands of elements efficiently

## Architecture Components

1. **SSE Hub**: Manages organization-scoped connections
2. **Connection Management**: Graceful registration/cleanup
3. **Event Broadcasting**: HTML payloads with semantic event names
4. **Originator Filtering**: Prevents action echo effects
5. **Dual Response System**: HTMX + SSE responses

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

## Quick Start

1. Set up SSE hub with connection management
2. Create HTML templates for updates
3. Add SSE wrapper elements with `hx-ext="sse"`
4. Implement dual response handlers
5. Generate unique originator IDs per browser tab

The pattern turns real-time sync from a complex state management problem into simple HTML broadcasting. The server becomes the single source of truth for how things should look, eliminating client-side UI complexity.