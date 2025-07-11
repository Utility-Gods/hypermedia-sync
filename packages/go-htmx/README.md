# 10,000 Checkboxes

An experiment pushing the boundaries of real-time hypermedia synchronization. How far can we take SSE + HTMX?

## The Vision

Forget WebSockets. Forget JSON APIs. Pure HTML over the wire, synchronized across browsers using Server-Sent Events. Every checkbox click broadcasts a tiny HTML fragment to all connected clients. No client-side state. No reconciliation. Just immediate DOM swaps.

10,000 checkboxes. Instant sync. Zero JavaScript logic.

## Run It

```bash
# Hot reload development
make dev

# Production
make build && ./bin/main
```

## Tech Stack

- **Go + Echo** - Blazing fast server
- **HTMX** - Hypermedia engine
- **SSE** - One-way real-time streams
- **Air** - Hot module reload

## The Pattern

1. User clicks checkbox
2. Server broadcasts HTML fragment to all clients (except originator)
3. HTMX swaps the exact DOM node
4. Everyone sees the same state instantly

No diffing. No virtual DOM. No state management. Just HTML flying through the tubes.

Port 8080. Open multiple tabs. Click around. Watch the magic.