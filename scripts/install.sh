#!/bin/bash
# rsyslox install script
# Installs rsyslox as a systemd service and starts the setup wizard.
#
# Usage:
#   sudo ./install.sh              # Full install
#   sudo ./install.sh --uninstall  # Remove rsyslox

set -euo pipefail

# ── Config ────────────────────────────────────────────────────────────────────
BINARY_NAME="rsyslox"
INSTALL_DIR="/opt/rsyslox"
CONFIG_DIR="/etc/rsyslox"
SERVICE_FILE="/etc/systemd/system/rsyslox.service"
SERVICE_USER="rsyslox"
SERVICE_GROUP="rsyslox"
DEFAULT_PORT=8000

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

info()    { echo -e "${BLUE}→${NC} $*"; }
success() { echo -e "${GREEN}✓${NC} $*"; }
warn()    { echo -e "${YELLOW}⚠${NC} $*"; }
error()   { echo -e "${RED}✗${NC} $*" >&2; }
header()  { echo -e "\n${BOLD}$*${NC}"; }

# ── Root check ────────────────────────────────────────────────────────────────
if [[ $EUID -ne 0 ]]; then
    error "This script must be run as root."
    echo "  Try: sudo ./install.sh"
    exit 1
fi

# ── Uninstall mode ────────────────────────────────────────────────────────────
if [[ "${1:-}" == "--uninstall" ]]; then
    header "Uninstalling rsyslox"

    if systemctl is-active --quiet rsyslox 2>/dev/null; then
        info "Stopping service…"
        systemctl stop rsyslox
    fi

    if systemctl is-enabled --quiet rsyslox 2>/dev/null; then
        info "Disabling service…"
        systemctl disable rsyslox
    fi

    [[ -f "$SERVICE_FILE" ]] && { rm -f "$SERVICE_FILE"; systemctl daemon-reload; success "Service removed"; }
    [[ -d "$INSTALL_DIR" ]] && { rm -rf "$INSTALL_DIR"; success "Binary removed ($INSTALL_DIR)"; }

    # Do NOT remove config — user data
    if [[ -d "$CONFIG_DIR" ]]; then
        warn "Configuration kept at $CONFIG_DIR — remove manually if no longer needed."
    fi

    if id "$SERVICE_USER" &>/dev/null; then
        userdel "$SERVICE_USER" 2>/dev/null || true
        success "System user '$SERVICE_USER' removed"
    fi

    success "rsyslox uninstalled."
    exit 0
fi

# ── Detect binary ─────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="$SCRIPT_DIR/$BINARY_NAME"

if [[ ! -f "$BINARY" ]]; then
    error "Binary not found: $BINARY"
    echo "  Make sure install.sh and the rsyslox binary are in the same directory."
    exit 1
fi

# ── Banner ────────────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}rsyslox Installer${NC}"
echo "══════════════════════════════════════"
echo ""

# ── System user ───────────────────────────────────────────────────────────────
header "Creating system user"

if ! id "$SERVICE_USER" &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin \
        --comment "rsyslox service account" "$SERVICE_USER"
    success "User '$SERVICE_USER' created"
else
    success "User '$SERVICE_USER' already exists"
fi

# ── Install binary ────────────────────────────────────────────────────────────
header "Installing binary"

mkdir -p "$INSTALL_DIR"
cp "$BINARY" "$INSTALL_DIR/$BINARY_NAME"
chmod 755 "$INSTALL_DIR/$BINARY_NAME"
chown root:root "$INSTALL_DIR/$BINARY_NAME"
success "Binary installed to $INSTALL_DIR/$BINARY_NAME"

# ── Config directory ──────────────────────────────────────────────────────────
header "Setting up configuration directory"

mkdir -p "$CONFIG_DIR"
# rsyslox user must be able to read config (written by setup wizard as root→rsyslox)
chown root:"$SERVICE_GROUP" "$CONFIG_DIR"
chmod 750 "$CONFIG_DIR"
success "Config directory: $CONFIG_DIR (root:$SERVICE_GROUP, 750)"

# ── Systemd service ───────────────────────────────────────────────────────────
header "Installing systemd service"

# Detect existing port from config if present
CONFIGURED_PORT=$DEFAULT_PORT
if [[ -f "$CONFIG_DIR/config.toml" ]]; then
    PORT_FROM_CONFIG=$(grep -oP 'port\s*=\s*\K\d+' "$CONFIG_DIR/config.toml" 2>/dev/null | head -1 || true)
    [[ -n "$PORT_FROM_CONFIG" ]] && CONFIGURED_PORT="$PORT_FROM_CONFIG"
fi

# Write systemd unit
cat > "$SERVICE_FILE" << UNIT
[Unit]
Description=rsyslox — syslog viewer and REST API
Documentation=https://github.com/phil-bot/rsyslox
After=network.target mysql.service mariadb.service
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_GROUP
ExecStart=$INSTALL_DIR/$BINARY_NAME
Restart=on-failure
RestartSec=5s

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$CONFIG_DIR
PrivateTmp=true

# Environment — override config path if needed
# Environment=RSYSLOX_CONFIG=/etc/rsyslox/config.toml

[Install]
WantedBy=multi-user.target
UNIT

chmod 644 "$SERVICE_FILE"
systemctl daemon-reload
success "Service installed: $SERVICE_FILE"

# ── Check for existing config ─────────────────────────────────────────────────
if [[ -f "$CONFIG_DIR/config.toml" ]]; then
    header "Existing configuration detected"
    success "Config found at $CONFIG_DIR/config.toml — skipping setup wizard."
    echo ""
    info "Starting rsyslox with existing configuration…"
    systemctl enable rsyslox
    systemctl restart rsyslox
    sleep 1

    if systemctl is-active --quiet rsyslox; then
        success "rsyslox is running."
        echo ""
        echo -e "  Open ${BLUE}http://localhost:${CONFIGURED_PORT}${NC} in your browser."
    else
        error "rsyslox failed to start. Check logs:"
        echo "  journalctl -u rsyslox -n 30 --no-pager"
    fi
    exit 0
fi

# ── First-run: start in setup mode ────────────────────────────────────────────
header "First-run setup"

info "Starting rsyslox in setup mode…"
systemctl enable rsyslox
systemctl start rsyslox
sleep 1

if ! systemctl is-active --quiet rsyslox; then
    error "rsyslox failed to start. Check logs:"
    echo "  journalctl -u rsyslox -n 30 --no-pager"
    exit 1
fi

success "rsyslox is running in setup mode."
echo ""

# Try to open browser automatically
SETUP_URL="http://localhost:${DEFAULT_PORT}"
if command -v xdg-open &>/dev/null && [[ -n "${DISPLAY:-}" ]]; then
    xdg-open "$SETUP_URL" 2>/dev/null &
elif command -v open &>/dev/null; then
    open "$SETUP_URL" 2>/dev/null &
fi

echo "══════════════════════════════════════════════════════"
echo -e "${BOLD}Installation complete!${NC}"
echo "══════════════════════════════════════════════════════"
echo ""
echo -e "  ${BOLD}Next step:${NC} Complete setup in your browser:"
echo ""
echo -e "    ${BLUE}${BOLD}http://$(hostname -I | awk '{print $1}'):${DEFAULT_PORT}${NC}"
echo ""
echo "  The setup wizard is reachable from any machine on your network."
echo "  Once setup is complete, rsyslox will require authentication."
echo ""
echo "  Useful commands:"
echo "    systemctl status rsyslox          # Service status"
echo "    journalctl -u rsyslox -f          # Live logs"
echo "    sudo ./install.sh --uninstall     # Remove rsyslox"
echo ""
