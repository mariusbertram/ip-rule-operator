- Double pulsating rings (outer & inner)
    - Intensive node glow effects
    - Pulsating connections with thickness changes
    - Flowing routing arrows with labels
    - Glowing TABLE 100/200 labels
    - Animated K8s badge with rotation and scaling
    - Pulsating title with color-changing glow

### 2. Documentation
- ✅ **`docs/README.md`** - Logo documentation
  - Design elements and color palette
  - Usage instructions
  - OLM Base64 encoding guide
  - PNG export commands

### 3. Helper Scripts
- ✅ **`docs/encode-logo.sh`** - Base64 encoding script (Linux/Mac/WSL)
- ✅ **`docs/encode-logo.ps1`** - Base64 encoding script (Windows PowerShell)

### 4. README Integration
- ✅ Logo added to main README
- ✅ Centered display with badges
- ✅ Professional layout

## 🎨 Design Concept

### Symbolism
- **Blue Circle**: Kubernetes Cluster
- **Golden Nodes**: Kubernetes Nodes/Services
- **White Lines**: Network connections
- **Dashed Rings**: Policy areas (CIDR)
- **Green Arrow**: Primary route (Table 100)
- **Red Arrow**: Secondary route (Table 200)

### Colors
- Kubernetes Blue: #326CE5
- Dark Blue: #1A4D8F  
- Gold: #FFD700
- Orange: #FFA500
- Green: #00FF7F
- Red: #FF6B6B

### 🎭 Animation Effects

The logos are fully animated to attract more attention:

#### Standard Logo (logo.svg)
- **Network Rings**: Pulsate and rotate (3-4s cycles)
- **Nodes**: Golden glow effect that gets stronger and weaker (2s cycles)
- **Connections**: Pulsating thickness and opacity (2s cycles, time-shifted)
- **Routing Arrows**: Flowing movement from left to right (1.5s cycles)
- **Text**: Glowing effect with green highlight (3s cycles)

#### Catalog Logo (logo-catalog.svg)
- **Outer Rings**: Double pulsating animations (4s cycles, time-shifted)
- **Nodes**: Intensive glow with orange highlight (2.5s cycles)
- **Connections**: Pulsating thickness from 2px to 3px (2s cycles)
- **Routing Arrows**: Flowing movement with opacity changes (2s cycles)
- **Route Labels**: Glowing labels "TABLE 100" and "TABLE 200" (2.5s cycles)
- **K8s Badge**: Rotation and scaling with opacity changes (4s cycles)
- **Title**: Pulsating glow from white to green (4s cycles)

#### Technical Details
- All animations use CSS keyframes within the SVG
- Smooth `ease-in-out` transitions for natural movements
- Time offsets (`animation-delay`) for coordinated effects
- No external dependencies - pure SVG+CSS
- Cross-browser compatibility (modern browsers)
- Animation loop is infinite (`infinite`)

#### Performance
- Lightweight: No JavaScript dependencies
- Hardware-accelerated: CSS animations use GPU
- Scalable: SVG format stays sharp at any size
- Base64-encodable: Works in OLM/OperatorHub as well

## 📋 Next Steps for OLM Integration

### 1. Create Base64 Encoding

```bash
# Linux/Mac/WSL
cd docs
./encode-logo.sh

# Windows PowerShell
cd docs
.\encode-logo.ps1
```

### 2. Update ClusterServiceVersion

Add the encoded logo to `config/manifests/bases/ip-rule-operator.clusterserviceversion.yaml`:

```yaml
spec:
  icon:
  - base64data: <BASE64_STRING_HERE>
    mediatype: image/svg+xml
```

### 3. Regenerate Bundle

```bash
make bundle VERSION=0.0.1
```

## ✅ Usage

### In GitHub README
The logo is automatically loaded from `docs/logo.svg`.

### In Documentation
```markdown
![IP Rule Operator Logo](../docs/logo.svg)
```

### For External Websites
Use the raw link:
```
https://raw.githubusercontent.com/mariusbertram/ip-rule-operator/main/docs/logo.svg
```

## 🖼️ Creating PNG Versions (Optional)

If PNG versions are needed:

```bash
# With ImageMagick (logo with transparent background)
convert -background none docs/logo.svg -resize 512x512 docs/logo-512.png
convert -background none docs/logo.svg -resize 256x256 docs/logo-256.png
convert -background none docs/logo.svg -resize 128x128 docs/logo-128.png
convert -background none docs/logo.svg -resize 64x64 docs/logo-64.png

# Catalog logo (with background)
convert docs/logo-catalog.svg -resize 256x256 docs/logo-catalog-256.png
```

## ✅ Done!

All logo assets have been successfully created and integrated into the README.

