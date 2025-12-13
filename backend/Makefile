.PHONY: dev test lint build clean migrate-up migrate-down

# Development with hot reload
dev:
	air

# Run tests
test:
	go test -v -race ./...

# Lint code
lint:
	golangci-lint run

# Build binary
build:
	go build -o bin/server ./cmd/server

# Clean build artifacts
clean:
	rm -rf bin/ tmp/

# Database migrations
migrate-up:
	@echo "Running migrations..."
	@for f in migrations/*.up.sql; do \
		echo "Applying $$f"; \
		psql "$$DATABASE_URL" -f "$$f"; \
	done

migrate-down:
	@echo "Rolling back migrations..."
	@for f in $$(ls -r migrations/*.down.sql); do \
		echo "Rolling back $$f"; \
		psql "$$DATABASE_URL" -f "$$f"; \
	done

# Generate JWT secret
gen-secret:
	@openssl rand -hex 32
