#!/bin/bash
# Encode logo for Operator Catalog CSV

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOGO_FILE="$SCRIPT_DIR/logo-catalog.svg"

if [ ! -f "$LOGO_FILE" ]; then
    echo "Error: $LOGO_FILE not found"
    exit 1
fi

echo "Encoding $LOGO_FILE to base64..."
echo ""
echo "Copy the following base64 string to your ClusterServiceVersion metadata:"
echo "============================================================================"
echo ""

if command -v base64 &> /dev/null; then
    # Linux/Mac
    base64 -w 0 "$LOGO_FILE"
elif command -v openssl &> /dev/null; then
    # Alternative mit openssl
    openssl base64 -A -in "$LOGO_FILE"
else
    echo "Error: Neither 'base64' nor 'openssl' command found"
    exit 1
fi

echo ""
echo ""
echo "============================================================================"
echo "Add this to your CSV file under spec.icon:"
echo ""
echo "  icon:"
echo "  - base64data: <paste-base64-here>"
echo "    mediatype: image/svg+xml"
echo ""

