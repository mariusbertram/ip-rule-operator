#!/bin/bash
# Script to add agent image to CSV relatedImages with digest resolution
# This script is safe and only modifies the relatedImages section

set -e

CSV_FILE="bundle/manifests/ip-rule-operator.clusterserviceversion.yaml"
AGENT_IMG="$1"
USE_DIGESTS="${2:-true}"

if [ -z "$AGENT_IMG" ]; then
    echo "Error: AGENT_IMG not provided"
    exit 1
fi

echo "Processing agent image: $AGENT_IMG"
echo "Use digests: $USE_DIGESTS"

# Function to get image digest
get_image_digest() {
    local image="$1"
    local digest=""

    echo "  Attempting to resolve digest for: $image"

    # Try skopeo first (works without pulling image)
    if command -v skopeo &> /dev/null; then
        echo "  Trying skopeo inspect..."
        digest=$(skopeo inspect --format='{{.Digest}}' docker://"${image}" 2>/dev/null || echo "")
        if [ -n "$digest" ]; then
            local registry_image="${image%:*}"
            echo "  ✓ Resolved via skopeo: ${registry_image}@${digest}"
            echo "${registry_image}@${digest}"
            return 0
        fi
    fi

    # Try docker inspect (requires image to be pulled)
    if command -v docker &> /dev/null; then
        echo "  Trying docker inspect..."
        digest=$(docker inspect --format='{{index .RepoDigests 0}}' "${image}" 2>/dev/null || echo "")
        if [ -n "$digest" ]; then
            echo "  ✓ Resolved via docker: $digest"
            echo "$digest"
            return 0
        fi
    fi

    # Try podman inspect (requires image to be pulled)
    if command -v podman &> /dev/null; then
        echo "  Trying podman inspect..."
        digest=$(podman inspect --format='{{index .RepoDigests 0}}' "${image}" 2>/dev/null || echo "")
        if [ -n "$digest" ]; then
            echo "  ✓ Resolved via podman: $digest"
            echo "$digest"
            return 0
        fi
    fi

    # If no digest found, return original image
    echo "  ⚠ Could not resolve digest, using original tag"
    echo "$image"
}

# Resolve agent image to digest if requested
FINAL_AGENT_IMG="$AGENT_IMG"
if [ "$USE_DIGESTS" = "true" ]; then
    echo ""
    echo "Resolving agent image digest..."
    AGENT_IMG_WITH_DIGEST=$(get_image_digest "$AGENT_IMG")
    if [[ "$AGENT_IMG_WITH_DIGEST" == *"@sha256:"* ]]; then
        FINAL_AGENT_IMG="$AGENT_IMG_WITH_DIGEST"
        echo ""
        echo "✓ Agent image resolved to digest"
    else
        echo ""
        echo "⚠ Using tag instead of digest"
        FINAL_AGENT_IMG="$AGENT_IMG"
    fi
fi

echo ""
echo "Final agent image: $FINAL_AGENT_IMG"
echo ""

# Check if CSV file exists
if [ ! -f "$CSV_FILE" ]; then
    echo "Error: CSV file not found: $CSV_FILE"
    exit 1
fi

# Check if relatedImages section exists
if ! grep -q "relatedImages:" "$CSV_FILE"; then
    echo "Error: relatedImages section not found in CSV"
    echo "This script should run after operator-sdk generates the bundle with USE_IMAGE_DIGESTS=true"
    exit 1
fi

# Check if agent image already exists in relatedImages section specifically
if sed -n '/relatedImages:/,/^  [a-z]/p' "$CSV_FILE" | grep -q "name: agent"; then
    echo "Agent image already present in relatedImages, skipping..."
else
    echo "Adding agent image to relatedImages..."
    # Use awk for safe in-place editing
    awk -v img="$FINAL_AGENT_IMG" '
    /name: manager/ && in_related {
        print
        print "    - image: " img
        print "      name: agent"
        next
    }
    /relatedImages:/ { in_related=1 }
    /^  [a-z]/ && !/relatedImages:/ { in_related=0 }
    { print }
    ' "$CSV_FILE" > "${CSV_FILE}.tmp" && mv "${CSV_FILE}.tmp" "$CSV_FILE"
    echo "✓ Agent image added to relatedImages"
fi

echo ""
echo "Verifying relatedImages section:"
sed -n '/relatedImages:/,/^  [a-z]/p' "$CSV_FILE" | head -n -1

