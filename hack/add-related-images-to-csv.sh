#!/bin/bash
# Script to add multiple images to CSV relatedImages with digest resolution
# Reads image list from a configuration file

set -e

CSV_FILE="bundle/manifests/ip-rule-operator.clusterserviceversion.yaml"
IMAGES_CONFIG="${1:-config/manifests/related-images.txt}"
USE_DIGESTS="${2:-true}"

echo "========================================"
echo "CSV Related Images Update Script"
echo "========================================"
echo "CSV File: $CSV_FILE"
echo "Images Config: $IMAGES_CONFIG"
echo "Use Digests: $USE_DIGESTS"
echo ""

# Check if CSV file exists
if [ ! -f "$CSV_FILE" ]; then
    echo "❌ Error: CSV file not found: $CSV_FILE"
    exit 1
fi

# Check if images config file exists
if [ ! -f "$IMAGES_CONFIG" ]; then
    echo "❌ Error: Images config file not found: $IMAGES_CONFIG"
    echo "   Please create the file with format: <name>:<image>"
    exit 1
fi

# Check if relatedImages section exists
if ! grep -q "relatedImages:" "$CSV_FILE"; then
    echo "❌ Error: relatedImages section not found in CSV"
    echo "   This script should run after operator-sdk generates the bundle with USE_IMAGE_DIGESTS=true"
    exit 1
fi

# Function to get image digest
get_image_digest() {
    local image="$1"
    local digest=""

    # Try skopeo first (works without pulling image)
    if command -v skopeo &> /dev/null; then
        digest=$(skopeo inspect --format='{{.Digest}}' docker://"${image}" 2>/dev/null || echo "")
        if [ -n "$digest" ]; then
            local registry_image="${image%:*}"
            echo "${registry_image}@${digest}"
            return 0
        fi
    fi

    # Try docker inspect (requires image to be pulled)
    if command -v docker &> /dev/null; then
        digest=$(docker inspect --format='{{index .RepoDigests 0}}' "${image}" 2>/dev/null || echo "")
        if [ -n "$digest" ]; then
            echo "$digest"
            return 0
        fi
    fi

    # Try podman inspect (requires image to be pulled)
    if command -v podman &> /dev/null; then
        digest=$(podman inspect --format='{{index .RepoDigests 0}}' "${image}" 2>/dev/null || echo "")
        if [ -n "$digest" ]; then
            echo "$digest"
            return 0
        fi
    fi

    # If no digest found, return original image
    echo "$image"
}

# Function to check if image already exists in relatedImages
image_exists_in_csv() {
    local name="$1"
    sed -n '/relatedImages:/,/^  [a-z]/p' "$CSV_FILE" | grep -q "name: $name"
}

# Function to add image to relatedImages
add_image_to_csv() {
    local name="$1"
    local image="$2"

    # Use awk for safe in-place editing - add after the last relatedImage entry
    awk -v name="$name" -v img="$image" '
    /relatedImages:/ {
        in_related=1
        print
        next
    }
    in_related && /^  [a-z]/ && !/^    / {
        # End of relatedImages section, insert before this line
        print "    - image: " img
        print "      name: " name
        in_related=0
    }
    { print }
    END {
        # If we are still in relatedImages at EOF, add at the end
        if (in_related) {
            print "    - image: " img
            print "      name: " name
        }
    }
    ' "$CSV_FILE" > "${CSV_FILE}.tmp"

    if [ $? -eq 0 ] && [ -f "${CSV_FILE}.tmp" ]; then
        mv "${CSV_FILE}.tmp" "$CSV_FILE"
        return 0
    else
        rm -f "${CSV_FILE}.tmp"
        return 1
    fi
}

# Read and process images from config file
images_added=0
images_skipped=0
images_failed=0

echo "Processing images from: $IMAGES_CONFIG"
echo "----------------------------------------"

while IFS=: read -r name image || [ -n "$name" ]; do
    # Skip empty lines and comments
    [[ -z "$name" || "$name" =~ ^[[:space:]]*# ]] && continue

    # Trim whitespace
    name=$(echo "$name" | xargs)
    image=$(echo "$image" | xargs)

    # Skip if name or image is empty
    if [ -z "$name" ] || [ -z "$image" ]; then
        continue
    fi

    echo ""
    echo "📦 Processing: $name"
    echo "   Image: $image"

    # Check if image already exists
    if image_exists_in_csv "$name"; then
        echo "   ⏭️  Already present in relatedImages, skipping..."
        ((images_skipped++))
        continue
    fi

    # Resolve digest if requested
    final_image="$image"
    if [ "$USE_DIGESTS" = "true" ]; then
        echo "   🔍 Resolving digest..."
        resolved_image=$(get_image_digest "$image")

        if [[ "$resolved_image" == *"@sha256:"* ]]; then
            final_image="$resolved_image"
            echo "   ✅ Resolved to digest: ${final_image##*@}"
        else
            echo "   ⚠️  Could not resolve digest, using tag"
            final_image="$image"
        fi
    fi

    # Add to CSV
    echo "   ➕ Adding to relatedImages..."
    if add_image_to_csv "$name" "$final_image"; then
        echo "   ✅ Successfully added"
        ((images_added++))
    else
        echo "   ❌ Failed to add"
        ((images_failed++))
    fi

done < "$IMAGES_CONFIG"

echo ""
echo "========================================"
echo "Summary"
echo "========================================"
echo "✅ Added:   $images_added"
echo "⏭️  Skipped: $images_skipped"
echo "❌ Failed:  $images_failed"
echo ""

if [ $images_failed -gt 0 ]; then
    echo "⚠️  Some images failed to add. Please check the errors above."
    exit 1
fi

# Verify and display relatedImages section
echo "Current relatedImages section:"
echo "----------------------------------------"
sed -n '/relatedImages:/,/^  [a-z]/p' "$CSV_FILE" | head -n -1

echo ""
echo "✅ Done!"

