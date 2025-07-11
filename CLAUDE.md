# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a repository for demonstrating **Hypermedia-Driven Real-Time Synchronization with SSE + HTMX** across different technology stacks. The project showcases implementations using various frameworks while maintaining HTMX as the constant client-side driver.

## Current Implementations

1. **go-htmx**: Go + Echo framework implementation (complete)
   - Full SSE support with real-time updates
   - Dark theme with orange/navy color scheme
   - 10,000 checkbox demo

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
- The `docs/overview.md` contains the complete architectural specification
- The pattern is language-agnostic and can be implemented in any server framework
- Focus on HTML as the universal interface for both initial loads and real-time updates

## When Implementing New Stacks

1. Set up organization-scoped SSE hubs
2. Implement dual response handlers (HTMX + SSE)
3. Create semantic template separation (*Updated vs *SSE)
4. Add originator ID tracking and filtering
5. Test with multiple browser tabs for real-time sync verification
