# Hypermedia Sync Experiments

A Go application demonstrating **Hypermedia-Driven Real-Time Synchronization** using Server-Sent Events (SSE) and HTMX. This project showcases various interactive experiments that maintain true hypermedia compliance while providing seamless real-time collaboration.

## 🎯 The Experiment

Just trying to push the limits of hypermedia-driven synchronization. No WebSockets. No JSON. Just Server-Sent Events broadcasting tiny HTML snippets to connected browsers. HTMX swaps them into the DOM. Instant sync.

The question: **How far can we take this?**

## 🔄 The Pattern

```
Click → Server → HTML fragment → SSE broadcast → DOM swap everywhere
```

That's it. The server is the single source of truth. Browsers are just dumb terminals that swap HTML. No client state. No reconciliation algorithms. Just immediate, surgical DOM updates.

## 🧪 Current Experiments

### 10,000 Checkboxes (`/experiments/checkboxes`)
Our flagship experiment syncs 10,000 checkboxes across browsers in real-time. Each click broadcasts ~50 bytes of HTML. Open it in multiple tabs. Click around. Watch them sync instantly.

## 🚀 Running the Experiments

```bash
go run main.go
# Visit http://localhost:8080
```

Or with Docker:
```bash
docker build -t hypermedia-sync .
docker run -p 8080:8080 hypermedia-sync
```

## 🏗️ Project Structure

```
hypermedia-sync/
├── main.go                 # Application entry point & routing
├── handlers/               # Core route handlers
├── sse/                   # SSE hub infrastructure  
├── experiments/           # Individual experiments
│   └── checkboxes/        # 10K checkboxes experiment
├── static/                # CSS, JS, assets
└── docs/                  # Architecture documentation
```

## 💡 Why This Matters

We've been building SPAs with complex state management for years. But what if we didn't need any of that? What if the server could just... tell browsers exactly what to display?

This isn't about building the next React. It's about exploring a radically simple alternative for real-time collaborative interfaces.

## 🔍 Key Features

- **Real-time Synchronization**: Changes instantly propagate to all connected users
- **Hypermedia Compliance**: Server controls all UI through HTML, not JSON
- **Modular Experiments**: Easy framework for adding new real-time patterns
- **Performance Optimized**: Only affected elements update
- **Connection Management**: Automatic cleanup and online user tracking

## 🎨 Design

Modern dark theme with vibrant orange accents, responsive CSS Grid layouts, and smooth animations that work across all experiment types.

## 🔮 Adding New Experiments

1. Create directory in `/experiments/[name]/`
2. Implement handlers with SSE integration
3. Add to experiments listing
4. Include comprehensive documentation
5. Test real-time sync across multiple tabs

## 📚 Deep Dive

See [docs/understanding.md](docs/understanding.md) for comprehensive architecture details and implementation patterns.

---

*Demonstrating the power of hypermedia-driven real-time architecture*