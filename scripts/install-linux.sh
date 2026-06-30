#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
INSTALL_DIR="${INSTALL_DIR:-/opt/raven-webmarket}"
SERVICE_USER="${SERVICE_USER:-raven}"

echo "=== Raven Webmarket — Linux Production Install ==="
echo "Install directory: $INSTALL_DIR"

if [ "$(id -u)" -ne 0 ]; then
  echo "Run as root: sudo bash scripts/install-linux.sh"
  exit 1
fi

if ! id "$SERVICE_USER" &>/dev/null; then
  useradd --system --home-dir "$INSTALL_DIR" --shell /usr/sbin/nologin "$SERVICE_USER"
fi

mkdir -p "$INSTALL_DIR"
rsync -a --exclude node_modules --exclude .git "$ROOT/" "$INSTALL_DIR/"
chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"

if [ ! -f "$INSTALL_DIR/.env" ]; then
  cp "$INSTALL_DIR/.env.example" "$INSTALL_DIR/.env"
  echo "Created $INSTALL_DIR/.env — edit before starting services."
fi

echo "Building API binary..."
cd "$INSTALL_DIR/backend"
sudo -u "$SERVICE_USER" go build -o raven-api ./cmd/server

echo "Building frontend..."
cd "$INSTALL_DIR/frontend"
sudo -u "$SERVICE_USER" npm ci
sudo -u "$SERVICE_USER" npm run build
cp -r .next/standalone/* "$INSTALL_DIR/frontend/"
cp -r .next/static "$INSTALL_DIR/frontend/.next/static"
cp -r public "$INSTALL_DIR/frontend/public" 2>/dev/null || true

echo "Installing systemd units..."
sed "s|/opt/raven-webmarket|$INSTALL_DIR|g" "$INSTALL_DIR/deploy/systemd/raven-api.service" > /etc/systemd/system/raven-api.service
sed "s|/opt/raven-webmarket|$INSTALL_DIR|g" "$INSTALL_DIR/deploy/systemd/raven-frontend.service" > /etc/systemd/system/raven-frontend.service
systemctl daemon-reload
systemctl enable raven-api raven-frontend

echo ""
echo "Done. Next steps:"
echo "  1. Edit $INSTALL_DIR/.env (set TRUST_CLOUDFLARE=true behind Cloudflare)"
echo "  2. Run: bash $INSTALL_DIR/scripts/migrate.sh"
echo "  3. Start: systemctl start raven-api raven-frontend"
echo "  4. Check: systemctl status raven-api raven-frontend"
echo "  5. Logs:  journalctl -u raven-api -f"
echo ""
echo "Services run in background — no terminal window required."
