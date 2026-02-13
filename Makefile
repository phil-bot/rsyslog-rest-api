.PHONY: build build-static clean install

# Get version from git tag or use dev
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -s -w -X main.Version=$(VERSION)

# Build the application
build:
	@echo "Building rsyslog-rest-api ($(VERSION))..."
	@mkdir -p build
	@go build -ldflags "$(LDFLAGS)" -o build/rsyslog-rest-api .
	@echo "✓ Build complete: ./build/rsyslog-rest-api"

# Build static binary (no libc dependency)
build-static:
	@echo "Building static binary ($(VERSION))..."
	@mkdir -p build
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o build/rsyslog-rest-api .
	@echo "✓ Static build complete: ./build/rsyslog-rest-api"

# Clean build artifacts
clean:
	@rm -rf build/
	@echo "✓ Cleaned"

# Install to /opt/rsyslog-rest-api
install: build-static
	@echo "Installing to /opt/rsyslog-rest-api..."
	@sudo mkdir -p /opt/rsyslog-rest-api
	@sudo cp build/rsyslog-rest-api /opt/rsyslog-rest-api/
	@sudo chmod +x /opt/rsyslog-rest-api/rsyslog-rest-api
	@[ ! -f /opt/rsyslog-rest-api/.env ] && sudo cp .env.example /opt/rsyslog-rest-api/.env || true
	@sudo cp rsyslog-rest-api.service /etc/systemd/system/
	@sudo systemctl daemon-reload
	@echo "✓ Installed (version: $(VERSION))"
	@echo ""
	@echo "IMPORTANT - Configuration required:"
	@echo "  1. Set database credentials: sudo nano /opt/rsyslog-rest-api/.env"
	@echo "     (DB_HOST, DB_NAME, DB_USER, DB_PASS)"
	@echo "  2. Set API key: API_KEY=\$$(openssl rand -hex 32)"
	@echo "  3. Secure config: sudo chmod 600 /opt/rsyslog-rest-api/.env"
	@echo "  4. Start: sudo systemctl enable --now rsyslog-rest-api"
