.PHONY: build build-static clean install

# Build the application
build:
	@echo "Building rsyslog-rest-api..."
	@mkdir -p build
	@go build -ldflags "-s -w" -o build/rsyslog-rest-api .
	@echo "✓ Build complete: ./build/rsyslog-rest-api"

# Build static binary (no libc dependency)
build-static:
	@echo "Building static binary..."
	@mkdir -p build
	@CGO_ENABLED=0 go build -ldflags "-s -w" -o build/rsyslog-rest-api .
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
	@echo "✓ Installed. Configure: sudo nano /opt/rsyslog-rest-api/.env"
