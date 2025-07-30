# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Warnings
- NEVER ADD COMMENTS TO THE CODE< AND IF YOU SEE A COMMENT IN THE CODE that is AI generated, remove it/ only remove comments that are in the CODE you are editing


## Project Overview

This is a Go application demonstrating **Hypermedia-Driven Real-Time Synchronization with SSE + HTMX** through various interactive experiments. The project showcases different real-time patterns while maintaining HTMX as the client-side driver and true hypermedia compliance.

## Core Understanding

For a comprehensive understanding of the pattern and architecture, see [docs/understanding.md](docs/understanding.md). This document merges all architectural knowledge and implementation details.

## Current Structure

The application is organized as a single Go service with modular experiments:

- **Main Application**: Echo framework with SSE hub infrastructure
- **Experiments Framework**: Organized experiments in `/experiments/` directory  
- **Current Experiments**:
  - `checkboxes/`: 10,000 synchronized checkboxes demonstration

## Lessons Learned

- **Cloudflare Workers Limitations**: SSE doesn't work well with Workers due to request duration limits and I/O isolation. Workers are designed for short-lived requests, not long-running connections.
- **Alternative Approaches**: For edge deployments, consider polling, WebSockets with Durable Objects, or external real-time services.


## Architecture Philosophy

The core principle is **true hypermedia compliance**: servers control all UI representation by sending complete HTML contexts rather than JSON data. This eliminates client-side UI logic and maintains a single source of truth on the server.

### Key Patterns

1. **Dual Response Pattern**: 
   - HTMX responses for action originators (rich interactions with toasts, modals)
   - SSE broadcasts for other clients (pure content updates)

2. **SSE Component Semantics**:
   - `*Updated` templates: HTMX interaction responses with full UI context
   - `*SSE` templates: Pure data representation for broadcasts
   - `*WithToast` templates: HTMX responses with notifications

3. **Organization-Scoped Connections**: SSE connections are grouped by organization to prevent cross-tenant data leaks

4. **Originator Filtering**: Uses `X-Originator-ID` headers to prevent echo effects

## Architecture Components

### SSE Hub Infrastructure
- Organization-based connection management
- Connection ID matching between SSE and HTMX requests
- Graceful cleanup and lifecycle management

### Event Broadcasting System
- Named HTML events matching `sse-swap` attributes
- Organization-scoped filtering
- Connection exclusion for originators

### Frontend Integration
- Persistent SSE connection wrappers that survive navigation
- Targeted content swapping with `hx-target`
- Unique originator ID generation per browser tab

## Mental Models

### Component Hierarchy
```
Page Component
├── HTMX Component (rich interactions: toasts, modals, navigation)
├── SSE Component (pure content: clean HTML, no side effects) 
├── Business Logic Component
└── Data Component
```

### Template Semantics
- **HTMX templates**: Complete interaction responses with side effects
- **SSE templates**: Pure state representation without side effects
- Different semantic responsibilities require separate templates

## Development Notes

- Each package is a complete, runnable implementation
- The pattern is language-agnostic and can be implemented in any server framework
- Focus on HTML as the universal interface for both initial loads and real-time updates

## Adding New Experiments

1. Create new directory in `/experiments/[experiment-name]/`
2. Implement handler.go with Echo route handlers
3. Add experiment to the main experiments list in `handlers/experiments.go`
4. Include proper README.md documentation
5. Follow SSE hub patterns for real-time features
6. Test with multiple browser tabs for sync verification

## Application Structure

```
hypermedia-sync/
├── main.go                 # Application entry point
├── handlers/               # Core route handlers
│   ├── experiments.go      # Experiments listing page
│   └── sse.go             # SSE and health endpoints
├── sse/                   # SSE hub infrastructure
│   └── hub.go
├── experiments/           # Individual experiments
│   └── checkboxes/        # 10K checkboxes experiment
│       ├── handler.go
│       └── README.md
├── static/                # CSS, JS, assets
└── Dockerfile
```
