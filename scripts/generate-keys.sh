#!/bin/bash
set -e

ROOT=$(pwd)

API_KEYS="$ROOT/api/src/license"
APP_KEYS="$ROOT/app/src/license"

mkdir -p "$API_KEYS"
mkdir -p "$APP_KEYS"

echo "🔐 Generating Ed25519 keys..."

openssl genpkey -algorithm Ed25519 -out "$API_KEYS/private.pem"

openssl pkey -in "$API_KEYS/private.pem" -pubout -out "$APP_KEYS/public.pem"

chmod 600 "$API_KEYS/private.pem"
chmod 644 "$APP_KEYS/public.pem"

echo "✅ Keys generated"

echo ""
echo "Private key:"
echo "$API_KEYS/private.pem"
echo ""
echo "Public key:"
echo "$APP_KEYS/public.pem"
