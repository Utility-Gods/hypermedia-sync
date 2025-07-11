# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a conceptual repository for demonstrating **Hypermedia-Driven Real-Time Synchronization with SSE + HTMX**. The project is currently in the documentation phase, focusing on architectural patterns and implementation strategies.

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

- This is currently a documentation-only repository with no executable code
- The `docs/overview.md` contains the complete architectural specification
- Implementation examples are provided in Go but the pattern is language-agnostic
- Focus on HTML as the universal interface for both initial loads and real-time updates

## Future Implementation

When implementing this pattern:
1. Set up organization-scoped SSE hubs
2. Implement dual response handlers (HTMX + SSE)
3. Create semantic template separation (*Updated vs *SSE)
4. Add originator ID tracking and filtering
5. Test with multiple browser tabs for real-time sync verification