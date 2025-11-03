---

## Comparison

| Feature | Standard Logo | Catalog Logo |
|---------|---------------|--------------|
| Size | 200x200px | 256x256px |
| Background | Transparent/Blue | Gradient with Rounded Corners |
| Effects | ✅ Animations | Drop Shadow + ✅ Animations |
| Details | Simple | Detailed with Labels |
| Usage | README, Docs | OperatorHub, Catalogs |
| **Animations** | **7 Effects** | **11 Effects** |

---

## 🎨 Animation Features

### Standard Logo (logo.svg)
- ✅ Pulsating network rings (3s cycles)
- ✅ Glowing nodes with glow effect (2s cycles)
- ✅ Pulsating connection lines (time-shifted)
- ✅ Flowing routing arrows (green & red)
- ✅ Glowing text with dynamic glow

### Catalog Logo (logo-catalog.svg)
- ✅ Double pulsating rings (outer & inner)
- ✅ Intensive node glow effects
- ✅ Pulsating connections with thickness changes
- ✅ Flowing routing arrows with opacity changes
- ✅ Glowing "TABLE 100" and "TABLE 200" labels
- ✅ Animated K8s badge (rotation & scaling)
- ✅ Pulsating title with color transitions

**All Animations:** Pure SVG+CSS, no JavaScript dependencies!

📚 **More Details:** See [ANIMATIONS.md](ANIMATIONS.md) for technical documentation

---

## Design Elements

### Color Palette

- 🔵 **Kubernetes Blue**: `#326CE5` - Primary color
- 🔵 **Dark Blue**: `#1A4D8F` - Accents
- 🟡 **Gold**: `#FFD700` - Network Nodes
- 🟠 **Orange**: `#FFA500` - Node Borders
- 🟢 **Green**: `#00FF7F` - Routing Table 100
- 🔴 **Red**: `#FF6B6B` - Routing Table 200
- ⚪ **White**: `#FFFFFF` - Text & Lines
- ⚫ **Light Gray**: `#E0E0E0` - Secondary Text

### Symbols

- 🔵 **Large Circle**: Kubernetes Cluster
- 🟡 **Golden Nodes**: Kubernetes Nodes/Services
- ⚪ **White Lines**: Network Connections
- ⚫ **Dashed Rings**: Policy Areas (CIDR)
- ➡️ **Green Arrow**: Primary Route (Table 100)
- ➡️ **Red Arrow**: Secondary Route (Table 200)
- 📝 **Text**: "IP RULE" & "OPERATOR"

---

## Quick Start

### Include Logo in README
```markdown
![IP Rule Operator](docs/logo.svg)
```

### Prepare Logo for OLM
```bash
# Linux/Mac/WSL
cd docs
./encode-logo.sh

# Windows PowerShell
cd docs
.\encode-logo.ps1
```

Then insert the Base64-encoded version into the ClusterServiceVersion.

---

<div align="center">
  <p><em>Created for the IP Rule Operator Project</em></p>
  <p>Apache 2.0 License • Copyright 2025 Marius Bertram</p>
</div>

