#!/bin/bash
# Script to update related-images.txt with current Makefile variables
# This ensures the config file stays in sync with Makefile settings

set -e

IMAGES_CONFIG="${1:-config/manifests/related-images.txt}"
AGENT_IMG="${2}"

if [ -z "$AGENT_IMG" ]; then
    echo "Error: AGENT_IMG not provided"
    echo "Usage: $0 <config-file> <agent-image>"
    exit 1
fi

echo "Updating related-images.txt..."
echo "Agent Image: $AGENT_IMG"

# Create or update the config file
cat > "$IMAGES_CONFIG" << EOF
# IP Rule Operator - Related Images Configuration
# Diese Datei listet alle zusätzlichen Container-Images auf,
# die in die ClusterServiceVersion unter spec.relatedImages aufgenommen werden sollen.
#
# Format: <image-name>:<image-url>
# - image-name: Name des Images in relatedImages (z.B. "agent", "init-container")
# - image-url: Vollständige Image-URL mit Tag (z.B. "ghcr.io/org/image:v1.0.0")
#
# Kommentare (Zeilen beginnend mit #) werden ignoriert
# Leere Zeilen werden ignoriert
#
# Diese Datei wird automatisch von 'make bundle' aktualisiert

# Agent DaemonSet Image
agent:${AGENT_IMG}

# Weitere Images können hier manuell hinzugefügt werden:
# init-container:ghcr.io/mariusbertram/init-tool:v1.0.0
# sidecar:ghcr.io/mariusbertram/sidecar:v2.0.0
EOF

echo "✅ Updated: $IMAGES_CONFIG"

