# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk --no-cache add wget && \
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

# Copy health check script
COPY healthcheck.sh /healthcheck.sh
RUN chmod +x /healthcheck.sh

# Set default port
ENV PORT=8080

# Expose port
EXPOSE 8080

# Health check - uses PORT env var to check correct port
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD /healthcheck.sh

# Run the application
CMD ["./main"]
