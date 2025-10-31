#!/usr/bin/env python3
"""
Script to add multiple images to CSV relatedImages with digest resolution
Reads image list from a configuration file and safely updates the CSV YAML
"""

import sys
import os
import subprocess
import re
from pathlib import Path

def get_image_digest(image, use_digests=True):
    """
    Resolve image tag to digest using skopeo, docker, or podman
    """
    if not use_digests:
        return image

    print(f"   🔍 Resolving digest...")

    # Try skopeo first (recommended)
    if os.system("command -v skopeo >/dev/null 2>&1") == 0:
        try:
            result = subprocess.run(
                ["skopeo", "inspect", f"docker://{image}"],
                capture_output=True,
                text=True,
                timeout=30
            )
            if result.returncode == 0:
                for line in result.stdout.split('\n'):
                    if '"Digest":' in line:
                        digest = line.split('"')[3]
                        registry_image = image.rsplit(':', 1)[0]
                        resolved = f"{registry_image}@{digest}"
                        print(f"   ✅ Resolved to digest: {digest}")
                        return resolved
        except Exception as e:
            print(f"   ⚠️  Skopeo failed: {e}")

    # Try docker inspect
    if os.system("command -v docker >/dev/null 2>&1") == 0:
        try:
            result = subprocess.run(
                ["docker", "inspect", "--format={{index .RepoDigests 0}}", image],
                capture_output=True,
                text=True,
                timeout=10
            )
            if result.returncode == 0 and result.stdout.strip():
                resolved = result.stdout.strip()
                if '@sha256:' in resolved:
                    print(f"   ✅ Resolved via docker")
                    return resolved
        except Exception as e:
            print(f"   ⚠️  Docker failed: {e}")

    # Try podman inspect
    if os.system("command -v podman >/dev/null 2>&1") == 0:
        try:
            result = subprocess.run(
                ["podman", "inspect", "--format={{index .RepoDigests 0}}", image],
                capture_output=True,
                text=True,
                timeout=10
            )
            if result.returncode == 0 and result.stdout.strip():
                resolved = result.stdout.strip()
                if '@sha256:' in resolved:
                    print(f"   ✅ Resolved via podman")
                    return resolved
        except Exception as e:
            print(f"   ⚠️  Podman failed: {e}")

    print(f"   ⚠️  Could not resolve digest, using tag")
    return image

def load_images_config(config_file):
    """
    Load image list from configuration file
    Format: name:image
    """
    images = []
    with open(config_file, 'r') as f:
        for line in f:
            line = line.strip()
            # Skip empty lines and comments
            if not line or line.startswith('#'):
                continue

            # Parse name:image format
            if ':' in line:
                parts = line.split(':', 1)
                if len(parts) == 2:
                    name = parts[0].strip()
                    image = parts[1].strip()
                    if name and image:
                        images.append((name, image))

    return images

def image_exists_in_csv(csv_content, name):
    """
    Check if image with given name already exists in relatedImages
    """
    in_related = False
    for line in csv_content:
        if 'relatedImages:' in line:
            in_related = True
        elif in_related and re.match(r'^  [a-z]', line) and not line.startswith('    '):
            in_related = False
        elif in_related and f'name: {name}' in line:
            return True
    return False

def add_image_to_csv(csv_file, name, image):
    """
    Add image to relatedImages section in CSV file
    Uses line-by-line manipulation for safety
    """
    with open(csv_file, 'r') as f:
        lines = f.readlines()

    # Find relatedImages section and insertion point
    related_idx = None
    insert_idx = None
    in_related = False

    for i, line in enumerate(lines):
        if 'relatedImages:' in line:
            related_idx = i
            in_related = True
        elif in_related and re.match(r'^  [a-z]', line) and not line.startswith('    '):
            # End of relatedImages section
            insert_idx = i
            break

    if related_idx is None:
        raise Exception("relatedImages section not found in CSV")

    # If we didn't find end of section, append at end
    if insert_idx is None:
        insert_idx = len(lines)

    # Create new image entry
    new_entry = [
        f"    - image: {image}\n",
        f"      name: {name}\n"
    ]

    # Insert at the right position
    lines = lines[:insert_idx] + new_entry + lines[insert_idx:]

    # Write back
    with open(csv_file, 'w') as f:
        f.writelines(lines)

def main():
    if len(sys.argv) < 2:
        print("Usage: add-related-images-to-csv.py <images-config> [use_digests]")
        sys.exit(1)

    images_config = sys.argv[1]
    use_digests = sys.argv[2].lower() == 'true' if len(sys.argv) > 2 else True
    csv_file = "bundle/manifests/ip-rule-operator.clusterserviceversion.yaml"

    print("=" * 50)
    print("CSV Related Images Update Script (Python)")
    print("=" * 50)
    print(f"CSV File: {csv_file}")
    print(f"Images Config: {images_config}")
    print(f"Use Digests: {use_digests}")
    print()

    # Check files exist
    if not os.path.exists(csv_file):
        print(f"❌ Error: CSV file not found: {csv_file}")
        sys.exit(1)

    if not os.path.exists(images_config):
        print(f"❌ Error: Images config file not found: {images_config}")
        sys.exit(1)

    # Load CSV content
    with open(csv_file, 'r') as f:
        csv_content = f.readlines()

    # Check if relatedImages exists
    if not any('relatedImages:' in line for line in csv_content):
        print("❌ Error: relatedImages section not found in CSV")
        print("   This script should run after operator-sdk generates the bundle")
        sys.exit(1)

    # Load images from config
    images = load_images_config(images_config)

    if not images:
        print("⚠️  No images found in config file")
        sys.exit(0)

    print(f"Processing {len(images)} image(s) from: {images_config}")
    print("-" * 50)

    images_added = 0
    images_skipped = 0
    images_failed = 0

    for name, image in images:
        print()
        print(f"📦 Processing: {name}")
        print(f"   Image: {image}")

        # Check if already exists
        if image_exists_in_csv(csv_content, name):
            print(f"   ⏭️  Already present in relatedImages, skipping...")
            images_skipped += 1
            continue

        # Resolve digest if requested
        final_image = get_image_digest(image, use_digests)

        # Add to CSV
        print(f"   ➕ Adding to relatedImages...")
        try:
            add_image_to_csv(csv_file, name, final_image)
            print(f"   ✅ Successfully added")
            images_added += 1

            # Reload CSV content for next iteration
            with open(csv_file, 'r') as f:
                csv_content = f.readlines()
        except Exception as e:
            print(f"   ❌ Failed to add: {e}")
            images_failed += 1

    print()
    print("=" * 50)
    print("Summary")
    print("=" * 50)
    print(f"✅ Added:   {images_added}")
    print(f"⏭️  Skipped: {images_skipped}")
    print(f"❌ Failed:  {images_failed}")
    print()

    if images_failed > 0:
        print("⚠️  Some images failed to add. Please check the errors above.")
        sys.exit(1)

    # Display current relatedImages section
    print("Current relatedImages section:")
    print("-" * 50)
    with open(csv_file, 'r') as f:
        in_related = False
        for line in f:
            if 'relatedImages:' in line:
                in_related = True
            elif in_related and re.match(r'^  [a-z]', line) and not line.startswith('    '):
                break

            if in_related:
                print(line.rstrip())

    print()
    print("✅ Done!")

if __name__ == "__main__":
    main()

