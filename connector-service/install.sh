#!/bin/bash
set -e

echo "=== EKF One-on-One Connector Installation ==="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (use sudo)"
  exit 1
fi

# Create user
if ! id -u ekf-connector >/dev/null 2>&1; then
  echo "Creating ekf-connector user..."
  useradd --system --no-create-home --shell /bin/false ekf-connector
fi

# Create installation directory
echo "Creating installation directory..."
mkdir -p /opt/ekf-connector/logs
chmod 755 /opt/ekf-connector
chmod 755 /opt/ekf-connector/logs

# Build binary
echo "Building connector..."
go build -o /opt/ekf-connector/connector ./cmd/connector

# Copy config
echo "Copying configuration..."
cp config.yaml /opt/ekf-connector/

# Copy .env if exists
if [ -f ".env" ]; then
  cp .env /opt/ekf-connector/.env
  chmod 600 /opt/ekf-connector/.env
else
  echo "WARNING: .env file not found. Please create /opt/ekf-connector/.env from .env.example"
  cp .env.example /opt/ekf-connector/.env
  chmod 600 /opt/ekf-connector/.env
fi

# Set ownership
chown -R ekf-connector:ekf-connector /opt/ekf-connector

# Install systemd service
echo "Installing systemd service..."
cp ekf-connector.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable ekf-connector

echo ""
echo "=== Installation Complete ==="
echo ""
echo "Next steps:"
echo "1. Edit /opt/ekf-connector/.env with your credentials"
echo "2. Start the service: sudo systemctl start ekf-connector"
echo "3. Check status: sudo systemctl status ekf-connector"
echo "4. View logs: sudo journalctl -u ekf-connector -f"
