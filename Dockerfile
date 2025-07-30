# Build stage
FROM golang:1.24-alpine AS builder

# Install templ, node/npm, and other build dependencies
RUN apk --no-cache add wget nodejs npm curl && \
    go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate templ files
RUN templ generate

# Build CSS using Taskfile
RUN npm install -g @go-task/task && task build

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates and wget for HTTPS and health checks
RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy static files from builder stage
COPY --from=builder /app/static ./static

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]