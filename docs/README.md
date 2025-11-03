![Catalog Logo](logo-catalog.svg)

## 🎨 Animation Features

Both logos use modern SVG+CSS animations for dynamic presentation:

- **Pulsating Network Rings**: Smooth size changes and opacity transitions
- **Glowing Nodes**: Dynamic glow effect on all network nodes
- **Flowing Routing Arrows**: Horizontal movement of green and red arrows
- **Pulsating Connections**: Time-shifted animation of network lines
- **Glowing Text**: Dynamic glow on all text elements
- **Animated K8s Badge** (Catalog only): Rotation and scaling

📚 **Technical Details**: See [ANIMATIONS.md](ANIMATIONS.md) for complete documentation of all animation effects.

## Design Elements

### Colors
- **Primary (Kubernetes Blue)**: `#326CE5`
- **Secondary (Dark Blue)**: `#1A4D8F`
- **Accent (Gold/Yellow)**: `#FFD700`, `#FFA500`
- **Routing Paths**: 
  - Green: `#00FF7F` (Primary Route/Table 100)
  - Red: `#FF6B6B` (Secondary Route/Table 200)
- **Text**: `#FFFFFF` (White), `#E0E0E0` (Light Gray)

### Symbolism
- **Network Nodes (Golden Circles)**: Represent Kubernetes Nodes and Services
- **Connection Lines**: Symbolize network connections
- **Dashed Rings**: Represent policy areas/CIDR ranges
- **Routing Arrows**: 
  - Green Arrow → Routing Table 100
  - Red Arrow → Routing Table 200
  - Show Policy-Based Routing

## Usage

### In Markdown (GitHub README)
```markdown
![IP Rule Operator](docs/logo.svg)
```

### In HTML
```html
<img src="docs/logo.svg" alt="IP Rule Operator" width="200"/>
```

### For OLM/Operator Catalog

The `logo-catalog.svg` file should be referenced in the bundle metadata:

**config/manifests/bases/ip-rule-operator.clusterserviceversion.yaml**:
```yaml
metadata:
  annotations:
    # ...
  name: ip-rule-operator.v0.0.1
spec:
  # ...
  icon:
  - base64data: <base64-encoded logo-catalog.svg>
    mediatype: image/svg+xml
```

#### Base64 Encoding for OLM

```bash
# Linux/Mac/WSL
base64 -w 0 docs/logo-catalog.svg

# Windows PowerShell
[Convert]::ToBase64String([IO.File]::ReadAllBytes("docs\logo-catalog.svg"))
```

## PNG Export (Optional)

If PNG versions are needed:

```bash
# With Inkscape
inkscape logo.svg --export-type=png --export-filename=logo.png --export-width=512 --export-height=512

# With ImageMagick
convert -background none logo.svg -resize 512x512 logo.png

# For Catalog (with background)
convert logo-catalog.svg -resize 256x256 logo-catalog.png
```

## Recommended Sizes

| Usage | Size | Format | File |
|-------|------|--------|------|
| GitHub README Header | 200x200px | SVG | logo.svg |
| OperatorHub Catalog | 256x256px | SVG/PNG | logo-catalog.svg |
| Documentation | 150-200px | SVG | logo.svg |
| Website/Blog (small) | 64x64px | PNG | logo-64.png |
| Website/Blog (medium) | 128x128px | PNG | logo-128.png |
| Website/Blog (large) | 512x512px | PNG | logo-512.png |
| Favicon | 32x32px | PNG/ICO | favicon.ico |

## License

The logos are part of the IP Rule Operator project and are subject to the Apache 2.0 License.

Copyright 2025 Marius Bertram.

