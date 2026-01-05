#!/usr/bin/env bash
#
# Generate self-signed certificates for Federation mTLS (local development only)
#
# Usage:
#   ./scripts/generate-federation-certs.sh [output-dir]
#
# Default output-dir: ./certs/federation
#
# WARNING: These are self-signed certificates for local development only.
# DO NOT use in production. Use proper certificate management in production.

set -euo pipefail

OUTPUT_DIR="${1:-./certs/federation}"
DAYS=365

echo "===================================="
echo "Federation mTLS Certificate Generator"
echo "===================================="
echo ""
echo "Output directory: $OUTPUT_DIR"
echo "Validity: $DAYS days"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Generate CA private key
echo "1/5: Generating CA private key..."
openssl genrsa -out "$OUTPUT_DIR/ca-key.pem" 4096 2>/dev/null

# Generate CA certificate
echo "2/5: Generating CA certificate..."
openssl req -new -x509 -days $DAYS -key "$OUTPUT_DIR/ca-key.pem" \
  -out "$OUTPUT_DIR/ca-cert.pem" \
  -subj "/C=US/ST=CA/L=Local/O=AgentOS Dev/OU=Federation/CN=AgentOS Federation CA" \
  2>/dev/null

# Generate server private key
echo "3/5: Generating server private key..."
openssl genrsa -out "$OUTPUT_DIR/server-key.pem" 4096 2>/dev/null

# Generate server certificate signing request
echo "4/5: Generating server certificate..."
openssl req -new -key "$OUTPUT_DIR/server-key.pem" \
  -out "$OUTPUT_DIR/server-csr.pem" \
  -subj "/C=US/ST=CA/L=Local/O=AgentOS Dev/OU=Federation/CN=federation.local" \
  2>/dev/null

# Sign server certificate with CA
openssl x509 -req -days $DAYS -in "$OUTPUT_DIR/server-csr.pem" \
  -CA "$OUTPUT_DIR/ca-cert.pem" -CAkey "$OUTPUT_DIR/ca-key.pem" -CAcreateserial \
  -out "$OUTPUT_DIR/server-cert.pem" \
  2>/dev/null

# Generate client private key
echo "5/5: Generating client certificate..."
openssl genrsa -out "$OUTPUT_DIR/client-key.pem" 4096 2>/dev/null

# Generate client certificate signing request
openssl req -new -key "$OUTPUT_DIR/client-key.pem" \
  -out "$OUTPUT_DIR/client-csr.pem" \
  -subj "/C=US/ST=CA/L=Local/O=AgentOS Dev/OU=Federation/CN=federation-client" \
  2>/dev/null

# Sign client certificate with CA
openssl x509 -req -days $DAYS -in "$OUTPUT_DIR/client-csr.pem" \
  -CA "$OUTPUT_DIR/ca-cert.pem" -CAkey "$OUTPUT_DIR/ca-key.pem" -CAcreateserial \
  -out "$OUTPUT_DIR/client-cert.pem" \
  2>/dev/null

# Cleanup CSR files
rm -f "$OUTPUT_DIR"/*.csr "$OUTPUT_DIR"/ca-cert.srl

echo ""
echo "✅ Certificates generated successfully!"
echo ""
echo "Generated files:"
echo "  - $OUTPUT_DIR/ca-cert.pem (CA certificate)"
echo "  - $OUTPUT_DIR/ca-key.pem (CA private key)"
echo "  - $OUTPUT_DIR/server-cert.pem (server certificate)"
echo "  - $OUTPUT_DIR/server-key.pem (server private key)"
echo "  - $OUTPUT_DIR/client-cert.pem (client certificate)"
echo "  - $OUTPUT_DIR/client-key.pem (client private key)"
echo ""
echo "===================================="
echo "Environment Configuration"
echo "===================================="
echo ""
echo "Add these to your environment or .env file:"
echo ""
echo "# Federation mTLS (server)"
echo "export AGENTOS_FED_REQUIRE_MTLS=true"
echo "export AGENTOS_FED_SERVER_CERT_FILE=$OUTPUT_DIR/server-cert.pem"
echo "export AGENTOS_FED_SERVER_KEY_FILE=$OUTPUT_DIR/server-key.pem"
echo "export AGENTOS_FED_CA_CERT_FILE=$OUTPUT_DIR/ca-cert.pem"
echo ""
echo "# Federation mTLS (client)"
echo "export AGENTOS_FED_CLIENT_CERT_FILE=$OUTPUT_DIR/client-cert.pem"
echo "export AGENTOS_FED_CLIENT_KEY_FILE=$OUTPUT_DIR/client-key.pem"
echo "# CA cert is the same for both client and server"
echo ""
echo "===================================="
echo ""
echo "WARNING: These are self-signed certificates for local development only."
echo "DO NOT use in production!"
echo ""
