.PHONY: build build-static clean install

# Get version from git tag or use dev
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -s -w -X main.Version=$(VERSION)

# Build the application
build:
	@echo "Building rsyslox ($(VERSION))..."
	@mkdir -p build
	@go build -ldflags "$(LDFLAGS)" -o build/rsyslox .
	@echo "✓ Build complete: ./build/rsyslox"

# Build static binary (no libc dependency)
build-static:
	@echo "Building static binary ($(VERSION))..."
	@mkdir -p build
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o build/rsyslox .
	@echo "✓ Static build complete: ./build/rsyslox"

# Clean build artifacts
clean:
	@rm -rf build/
	@echo "✓ Cleaned"

# Install to /opt/rsyslox
install: build-static
	@echo "Installing to /opt/rsyslox..."
	@sudo mkdir -p /opt/rsyslox
	@sudo cp build/rsyslox /opt/rsyslox/
	@sudo chmod +x /opt/rsyslox/rsyslox
	@[ ! -f /opt/rsyslox/.env ] && sudo cp .env.example /opt/rsyslox/.env || true
	@sudo cp rsyslox.service /etc/systemd/system/
	@sudo systemctl daemon-reload
	@echo "✓ Installed (version: $(VERSION))"
	@echo ""
	@echo "IMPORTANT - Configuration required:"
	@echo "  1. Set database credentials: sudo nano /opt/rsyslox/.env"
	@echo "     (DB_HOST, DB_NAME, DB_USER, DB_PASS)"
	@echo "  2. Set API key: API_KEY=\$$(openssl rand -hex 32)"
	@echo "  3. Secure config: sudo chmod 600 /opt/rsyslox/.env"
	@echo "  4. Start: sudo systemctl enable --now rsyslox"
