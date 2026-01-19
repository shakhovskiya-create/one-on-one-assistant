#!/bin/bash
# Auto-setup script for One-on-One Assistant server
# Run this on fresh Ubuntu 22.04/24.04 server

set -e

echo "🚀 Setting up One-on-One Assistant Server"
echo "=========================================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ Please run as root (sudo bash setup-server.sh)"
    exit 1
fi

# Update system
echo "📦 Updating system packages..."
apt update && apt upgrade -y

# Install Docker
echo "🐳 Installing Docker..."
apt install -y apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt update
apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Install Docker Compose standalone
echo "📦 Installing Docker Compose..."
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Install Nginx
echo "🌐 Installing Nginx..."
apt install -y nginx

# Install Git
echo "📚 Installing Git..."
apt install -y git

# Install PostgreSQL client tools
echo "🐘 Installing PostgreSQL client..."
apt install -y postgresql-client

# Install useful tools
echo "🛠️  Installing utilities..."
apt install -y htop curl wget vim nano net-tools

# Enable and start Docker
systemctl enable docker
systemctl start docker

# Create app directory
echo "📁 Creating application directory..."
mkdir -p /opt/one-on-one
mkdir -p /opt/one-on-one/data/postgres
mkdir -p /opt/one-on-one/backups
mkdir -p /opt/one-on-one/logs

# Set permissions
chown -R root:root /opt/one-on-one

echo ""
echo "✅ Server setup complete!"
echo ""
echo "Next steps:"
echo "1. Clone the repository to /opt/one-on-one"
echo "2. Configure environment variables"
echo "3. Run docker-compose up -d"
echo ""
echo "Server info:"
echo "- Docker: $(docker --version)"
echo "- Docker Compose: $(docker-compose --version)"
echo "- Nginx: $(nginx -v 2>&1)"
echo ""
