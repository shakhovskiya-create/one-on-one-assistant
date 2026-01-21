#!/bin/bash

# Generate self-signed SSL certificate for development/testing
# For production, use Let's Encrypt or corporate certificates

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SSL_DIR="$PROJECT_DIR/nginx/ssl"

# Default values
DOMAIN="${1:-localhost}"
DAYS="${2:-365}"

echo "=== SSL Certificate Generator ==="
echo "Domain: $DOMAIN"
echo "Validity: $DAYS days"
echo "Output: $SSL_DIR"
echo ""

# Create SSL directory
mkdir -p "$SSL_DIR"

# Generate private key
echo "Generating private key..."
openssl genrsa -out "$SSL_DIR/key.pem" 2048

# Generate certificate signing request (CSR)
echo "Generating CSR..."
openssl req -new \
    -key "$SSL_DIR/key.pem" \
    -out "$SSL_DIR/cert.csr" \
    -subj "/CN=$DOMAIN/O=EKF Hub/C=RU"

# Generate self-signed certificate
echo "Generating self-signed certificate..."
openssl x509 -req \
    -days "$DAYS" \
    -in "$SSL_DIR/cert.csr" \
    -signkey "$SSL_DIR/key.pem" \
    -out "$SSL_DIR/cert.pem" \
    -extfile <(printf "subjectAltName=DNS:$DOMAIN,DNS:www.$DOMAIN,IP:127.0.0.1")

# Clean up CSR
rm -f "$SSL_DIR/cert.csr"

# Set permissions
chmod 600 "$SSL_DIR/key.pem"
chmod 644 "$SSL_DIR/cert.pem"

echo ""
echo "=== Certificate generated successfully ==="
echo ""
echo "Files created:"
echo "  - $SSL_DIR/cert.pem (certificate)"
echo "  - $SSL_DIR/key.pem (private key)"
echo ""
echo "To use with docker-compose-ssl.yml:"
echo "  docker-compose -f docker-compose-ssl.yml up -d"
echo ""
echo "WARNING: This is a self-signed certificate for development only!"
echo "For production, use Let's Encrypt or corporate certificates."
echo ""
echo "To install Let's Encrypt certificate:"
echo "  1. Install certbot: apt install certbot"
echo "  2. Generate certificate: certbot certonly --standalone -d $DOMAIN"
echo "  3. Copy certificates:"
echo "     cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem $SSL_DIR/cert.pem"
echo "     cp /etc/letsencrypt/live/$DOMAIN/privkey.pem $SSL_DIR/key.pem"
