#!/bin/bash
set -e

CERT_DIR="$(dirname "$0")/../certs"

# Create certs directory if not exists
mkdir -p "$CERT_DIR"

# Check if mkcert is installed
if ! command -v mkcert &> /dev/null; then
    echo "Error: mkcert is not installed. Please install it first:"
    echo "  brew install mkcert"
    exit 1
fi

# Install local CA if not already done
mkcert -install

# Generate certificates
cd "$CERT_DIR"
mkcert localhost 127.0.0.1 ::1

echo "Certificates generated in $CERT_DIR"
ls -la "$CERT_DIR"
