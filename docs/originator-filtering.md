# Originator Filtering with Server-Generated IDs

## Overview

The simplest approach to prevent duplicate updates is to generate unique originator IDs on the server and use them directly as SSE connection identifiers.

## How It Works

### 1. Server Generates Unique ID
When serving the initial page, the server generates a unique ID for that page load:

```go
// Generate unique originator ID for this page load
originatorID := fmt.Sprintf("page-%d-%d", time.Now().UnixNano(), rand.Intn(1000000))
```

### 2. ID Embedded in HTML
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

### 3. SSE Connection Uses ID
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

### 4. Broadcasting Excludes Originator
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

## Complete Flow

1. **Page Load**: Server generates unique ID, embeds in HTML
2. **SSE Connect**: Browser connects with `?originator={id}`
3. **User Action**: HTMX sends request with `X-Originator-ID` header
4. **Server Response**: 
   - Returns 204 No Content to originator (with `hx-swap="none"`)
   - Broadcasts HTML to all SSE connections except originator
5. **Result**: Clean updates without duplicates or unnecessary data transfer

## Benefits

- **Simplicity**: Direct ID matching, no complex topic systems
- **Server Control**: IDs generated server-side, no client coordination
- **Clean HTML**: SSE attributes work directly without JavaScript setup
- **Predictable**: Each page load gets one unique ID for its lifetime

## Code Example

```go
// Handle user action
func toggleHandler(c echo.Context) error {
    // Get originator ID from request
    originatorID := c.Request().Header.Get("X-Originator-ID")
    
    // Update state
    updateCheckbox(id)
    
    // Generate HTML
    html := generateCheckboxGridHTML()
    
    // Broadcast to all except originator
    hub.broadcast <- Event{
        Name:      "content-updated",
        Data:      html,
        ExcludeID: originatorID,
    }
    
    // Return no content (hx-swap="none")
    return c.NoContent(204)
}
```

This approach maintains pure hypermedia principles while providing clean originator filtering.