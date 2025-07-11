.PHONY: dev run build clean air-install

# Development with hot reload
dev:
	@if ! command -v air > /dev/null; then \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
	fi
	air

# Run without hot reload
run:
	go run main.go

# Build the application
build:
	go build -o bin/main main.go

# Clean build artifacts
clean:
	rm -rf tmp bin

# Install air globally
air-install:
	go install github.com/air-verse/air@latest