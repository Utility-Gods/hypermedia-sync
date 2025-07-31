# Understanding Hypermedia-Driven Real-Time Sync

## The Big Idea

Instead of sending JSON data and letting clients figure out display, send complete HTML from the server. Real-time updates become simple DOM replacements.

**Traditional:** Action → JSON → Client Processing → DOM Updates  
**Hypermedia:** Action → HTML → Direct DOM Replacement

## 🎯 Main Pattern: Pure Hypermedia + SSE

**Philosophy**: Server controls all state, pushes updates via SSE. No client-side state management.

### How It Works

1. **User clicks checkbox** → HTMX sends POST to server
2. **Server updates state** → Calculates new checkbox + counter state  
3. **Server broadcasts updates** → SSE pushes HTML to all clients
4. **All clients update** → Checkbox and counter sync in real-time

### Server Side (Go)

```go
func ToggleHandler() {
    // Update server state
    checkboxes[id] = !checkboxes[id]
    
    // Broadcast checkbox change
    hub.Broadcast(sse.Event{
        Name: fmt.Sprintf("checkbox-%d-updated", id),
        Data: renderCheckboxHTML(id),
    })
    
    // Broadcast counter update 
    totalChecked := countChecked(checkboxes)
    hub.Broadcast(sse.Event{
        Name: "counter-updated",
        Data: fmt.Sprintf("%d checked", totalChecked),
    })
}
```

### Client Side (HTML)

```html
<!-- Counter updates via SSE -->
<span sse-swap="counter-updated" hx-target="this">
    {{ initialCount }} checked
</span>

<!-- Checkbox triggers server change -->
<input type="checkbox" 
       hx-post="/toggle/1"
       hx-swap="outerHTML" 
       hx-target="this" />
```

### JavaScript (Minimal)

```javascript
// Only for adding originator IDs to requests
document.addEventListener('htmx:configRequest', function(evt) {
    evt.detail.headers['X-Originator-ID'] = window.originatorId;
});
```

## SSE Target Inheritance Fix

**Problem**: SSE elements inherit `hx-target` from parent `hx-boost` wrappers.

**Solution**: Always add explicit `hx-target="this"` to SSE elements.

```html
<!-- ❌ Wrong: inherits parent's hx-target -->
<div sse-swap="my-event" hx-swap="innerHTML">

<!-- ✅ Correct: explicit target -->
<div sse-swap="my-event" hx-swap="innerHTML" hx-target="this">
```

## Originator Filtering

Prevents duplicate updates by excluding the action originator from SSE broadcasts.

### Implementation

1. **Server generates unique ID** per page load
2. **SSE connection** uses ID: `/events?originator={id}`
3. **HTMX requests** include ID in `X-Originator-ID` header
4. **Broadcasting excludes** the originator connection

```go
// Handle action + broadcast to others
func toggleHandler(c echo.Context) error {
    originatorID := c.Request().Header.Get("X-Originator-ID")
    
    // Update state + broadcast to all except originator
    hub.Broadcast(Event{
        Name:      "item-updated",
        Data:      renderHTML(),
        ExcludeID: originatorID,
    })
    
    return c.NoContent(204)
}
```

## Summary

**Main Pattern**: Server state + SSE updates + minimal JavaScript  
**SSE Fix**: Always use `hx-target="this"`  
**Key Insight**: HTML broadcasting eliminates client-side complexity

The server becomes the single source of truth for how things should look, turning real-time sync from a complex state management problem into simple HTML broadcasting.
