# Go HTMX Implementation

Real-time hypermedia synchronization using Go, Echo, HTMX, and Server-Sent Events.

## Quick Start

### Development with Hot Reload

```bash
# Install dependencies
go mod download

# Run with hot reload (installs air automatically if needed)
make dev

# Or manually install air and run
go install github.com/air-verse/air@latest
air
```

### Standard Run

```bash
# Run without hot reload
make run

# Or directly
go run main.go
```

### Build

```bash
# Build binary
make build

# Run the binary
./bin/main
```

## Features

- Hot module reload using Air
- 10,000 real-time synchronized checkboxes
- Server-Sent Events for real-time updates
- Originator filtering to prevent echo effects
- External CSS for easy styling updates
- Dark theme with orange/navy color scheme

## File Structure

```
.
├── main.go              # Main application
├── static/
│   ├── css/
│   │   └── styles.css   # External styles
│   └── js/
│       ├── htmx.js      # HTMX library
│       └── sse.js       # SSE extension
├── .air.toml            # Air configuration
├── Makefile             # Build commands
└── tmp/                 # Air temporary files (gitignored)
```

## Hot Reload Configuration

The `.air.toml` file configures:
- Watch for changes in `.go`, `.html`, `.css`, `.js` files
- Automatic rebuild on file changes
- Colored output for better development experience
- Excludes `tmp/` directory from watching

## Development Tips

1. CSS changes are reflected immediately on page refresh
2. Go code changes trigger automatic rebuild and restart
3. Use `make clean` to remove build artifacts
4. The server runs on port 8080 by default