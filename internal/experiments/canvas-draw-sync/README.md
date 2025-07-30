# Canvas Draw Sync Experiment

A collaborative real-time drawing canvas showcasing hypermedia-driven synchronization with Server-Sent Events (SSE) and HTMX.

## Overview

This experiment demonstrates how multiple users can collaborate on a shared drawing canvas in real-time without using WebSockets or JSON APIs. All drawing operations are broadcast as SVG HTML fragments using SSE, maintaining true hypermedia compliance.

## Key Features

- **Real-time Collaboration**: Drawing strokes, shapes, and text appear instantly across all connected users
- **Multiple Drawing Tools**: Pen, rectangle, circle, and text tools with customizable colors and brush sizes
- **Hypermedia Compliance**: Server sends complete SVG HTML representations, not JSON data
- **Originator Filtering**: Users don't receive echoes of their own drawing actions
- **Online User Counter**: Live count of connected collaborative users
- **Canvas Management**: Real-time canvas clearing synchronized across all users
- **Immediate Visual Feedback**: Local drawing appears instantly while syncing to others

## Architecture

### SSE (Server-Sent Events)
- Long-lived HTTP connections for real-time drawing synchronization
- SVG HTML fragments broadcast to all connected clients
- Custom event handlers for proper SVG namespace management

### HTMX Integration
- `hx-post` for drawing actions and canvas operations
- Custom SSE message handlers for SVG element processing
- `X-Originator-ID` headers to prevent drawing echo effects

### Canvas State Management
- Global canvas state with thread-safe operations using mutexes
- In-memory storage of drawing elements with metadata
- SVG-based rendering for scalable graphics

## Implementation Details

### Dual Response Pattern
1. **Immediate Local Feedback**: JavaScript creates SVG elements instantly for responsive drawing
2. **SSE Broadcast**: Drawing data sent to all other connected users via server

### Template Structure
- SVG canvas with drawing tools and color picker
- Individual drawing element templates for real-time updates
- Canvas clearing and state management templates

### Drawing Tools
- **Pen Tool**: Freehand drawing with path elements
- **Rectangle Tool**: Click-to-place rectangular shapes
- **Circle Tool**: Click-to-place circular shapes  
- **Text Tool**: Click-to-place text elements with prompt input

## Technical Stack

- **Backend**: Go with Echo framework and mutex-protected state
- **Frontend**: HTMX + Server-Sent Events + SVG manipulation
- **Graphics**: SVG-based drawing with namespace-aware DOM manipulation
- **Styling**: Tailwind CSS with responsive canvas layout
- **Real-time**: SSE hub with connection management and event broadcasting

## Usage

1. Open the experiment in multiple browser tabs or different browsers
2. Select a drawing tool and color from the toolbar
3. Draw, create shapes, or add text on the canvas
4. Observe real-time updates appearing in all other connected sessions
5. Use "Clear Canvas" to reset the drawing for all users
6. Monitor the online user counter

This experiment showcases advanced hypermedia-driven real-time collaboration where complex interactive graphics are synchronized purely through HTML fragments, demonstrating the power of server-controlled UI state for collaborative applications.