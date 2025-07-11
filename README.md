# Hypermedia Sync

## The Experiment

Just trying to push the limits of hypermedia-driven synchronization. No WebSockets. No JSON. Just Server-Sent Events broadcasting tiny HTML snippets to connected browsers. HTMX swaps them into the DOM. Instant sync.

The question: **How far can we take this?**

## The Pattern

```
Click → Server → HTML fragment → SSE broadcast → DOM swap everywhere
```

That's it. The server is the single source of truth. Browsers are just dumb terminals that swap HTML. No client state. No reconciliation algorithms. Just immediate, surgical DOM updates.

## Current Test: 10,000 Checkboxes

Our Go implementation syncs 10,000 checkboxes across browsers in real-time. Each click broadcasts ~50 bytes of HTML. Open it in multiple tabs. Click around. Watch them sync instantly.

```bash
cd packages/go-htmx
make dev
```

## Why This Matters

We've been building SPAs with complex state management for years. But what if we didn't need any of that? What if the server could just... tell browsers exactly what to display?

This isn't about building the next React. It's about exploring a radically simple alternative for real-time collaborative interfaces.

## Architecture Notes

See [docs/understanding.md](docs/understanding.md) for the deep dive on how this actually works.
