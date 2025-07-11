# Hypermedia Sync Experiments

## Overview

This repository explores **hypermedia-driven real-time synchronization patterns** using different technology stacks while maintaining HTMX as the constant client-side driver. The goal is to demonstrate how various server-side technologies can implement the same elegant pattern: servers control all UI representation by sending complete HTML contexts rather than JSON data.

## Core Principle

**True Hypermedia Compliance**: Eliminate client-side UI logic by having servers send complete HTML representations. This creates a single source of truth on the server and simplifies real-time synchronization to simple DOM replacements.

### The Pattern

1. **User Action** → Server receives HTMX request
2. **Server Processing** → Update state, generate fresh HTML
3. **Dual Response**:
   - Return 204 No Content to action originator
   - Broadcast HTML updates via SSE to all other clients
4. **Result** → All browsers show identical UI instantly

## Architecture Patterns

### 1. Dual Response Pattern

- **HTMX responses** for action originators (can include toasts, modals, rich interactions)
- **SSE broadcasts** for other clients (pure content updates)

### 2. Originator Filtering

- Each browser tab gets a unique ID
- Server excludes originator from SSE broadcasts
- Prevents echo effects and duplicate updates

### 3. Targeted Updates

- Broadcast minimal HTML for specific elements
- Each element listens for its own SSE event
- Efficient DOM updates (50 bytes vs 2KB)

### 4. Pure Hypermedia

- HTML-only responses, no JSON
- Server controls all UI representation
- Zero client-side state management

## Implementations

### ✅ go-htmx

- **Stack**: Go + Echo + HTMX + SSE
- **Features**:
  - Real-time updates with Server-Sent Events
  - Goroutines for concurrent connection management
  - Thread-safe state with sync.Map
  - 10,000 checkbox demo
  - Dark theme with orange accents
- **Status**: Complete

## Key Benefits

1. **Simple Mental Model**: Server controls all UI, clients just swap HTML
2. **Zero Client Logic**: No JSON parsing, no state management
3. **Framework Agnostic**: Works with any server that generates HTML
4. **Easy Debugging**: Inspect HTML directly
5. **Consistent UI**: All clients see identical representation

## When to Use This Pattern

### ✅ Great for:

- Admin dashboards with live updates
- Collaborative editing interfaces
- Real-time monitoring systems
- Team management tools
- Live polls and surveys

### ❌ Consider alternatives for:

- High-frequency updates (>10/second)
- Offline-first applications
- Complex client animations
- Real-time games
- Mobile apps with poor connectivity

## Getting Started

Each implementation has its own README with specific setup instructions. Start with any implementation to see the pattern in action:

```bash
# Go implementation
cd packages/go-htmx
docker build -t go-htmx .
docker run -p 8080:8080 go-htmx

# Or run directly with Go
go run main.go
```

## Documentation

- [Architecture Overview](docs/overview.md) - Detailed pattern explanation
- [Originator Filtering](docs/originator-filtering.md) - Implementation details

