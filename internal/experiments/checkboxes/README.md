# 10,000 Checkboxes Experiment

A real-time synchronized checkbox demonstration showcasing hypermedia-driven state management with Server-Sent Events (SSE) and HTMX.

## Overview

This experiment demonstrates how 10,000 checkboxes can be synchronized in real-time across multiple browser sessions without using WebSockets or JSON APIs. All state changes are broadcast as HTML fragments using SSE, maintaining the hypermedia principle.

## Key Features

- **Real-time Synchronization**: Changes made by one user are instantly visible to all connected users
- **Hypermedia Compliance**: Server sends complete HTML representations, not JSON data
- **Originator Filtering**: Users don't receive echoes of their own actions
- **Online User Counter**: Live count of connected users
- **Optimized Updates**: Only the affected checkbox is updated, not the entire grid

## Architecture

### SSE (Server-Sent Events)
- Long-lived HTTP connections for real-time updates
- HTML fragments broadcast to all connected clients
- Organization-scoped connections prevent cross-tenant data leaks

### HTMX Integration
- `hx-post` for checkbox toggle actions
- `sse-swap` attributes for targeted content replacement
- `X-Originator-ID` headers to prevent echo effects

### State Management
- Global checkbox state maintained on server
- Mutex-protected concurrent access
- In-memory storage for demonstration purposes

## Implementation Details

### Dual Response Pattern
1. **HTMX Response**: Immediate UI feedback for the originating user
2. **SSE Broadcast**: Update notification sent to all other connected users

### Template Structure
- Main page template with embedded HTMX and SSE configuration
- Individual checkbox HTML generation for updates
- Online user counter updates

## Technical Stack

- **Backend**: Go with Echo framework
- **Frontend**: HTMX + Server-Sent Events
- **Styling**: Custom CSS with CSS Grid for responsive layout
- **Real-time**: SSE hub with connection management

## Usage

1. Open the experiment in multiple browser tabs or different browsers
2. Toggle any checkbox in one tab
3. Observe real-time updates in all other tabs
4. Monitor the online user counter

This experiment showcases the power of hypermedia-driven architecture where the server maintains complete control over UI representation while providing seamless real-time collaboration.