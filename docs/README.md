# IP Rule Operator - Logos und Assets

Dieses Verzeichnis enthält die visuellen Assets für den IP Rule Operator.

> ✨ **NEU:** Alle Logos sind jetzt vollständig animiert für mehr Aufmerksamkeit und eine professionellere Darstellung!

## Logo-Versionen

### 1. Standard-Logo (`logo.svg`) 🎬
- **Verwendung**: README, Dokumentation, Präsentationen
- **Format**: SVG (vektorbasiert, skalierbar)
- **Größe**: 200x200px (Standardgröße)
- **Hintergrund**: Transparent/Kubernetes-Blau
- **Animationen**: 7 verschiedene Effekte (pulsierende Ringe, leuchtende Nodes, fließende Pfeile)

![Standard Logo](logo.svg)

### 2. Catalog-Logo (`logo-catalog.svg`) 🎬
- **Verwendung**: OpenShift OperatorHub, Operator Catalogs
- **Format**: SVG mit Hintergrund
- **Größe**: 256x256px
- **Hintergrund**: Gradient (Kubernetes-Blau)
- **Features**: Abgerundete Ecken, Drop-Shadow, detaillierter
- **Animationen**: 11 verschiedene Effekte (inkl. animiertem K8s-Badge, leuchtende Labels)

![Catalog Logo](logo-catalog.svg)

## 🎬 Animations-Features

Beide Logos nutzen moderne SVG+CSS Animationen für eine dynamische Darstellung:

- **Pulsierende Netzwerkringe**: Sanfte Größenänderung und Opacity-Wechsel
- **Leuchtende Nodes**: Dynamischer Glow-Effekt auf allen Netzwerk-Knoten
- **Fließende Routing-Pfeile**: Horizontale Bewegung der grünen und roten Pfeile
- **Pulsierende Verbindungen**: Zeitversetzte Animation der Netzwerk-Linien
- **Leuchtender Text**: Dynamischer Glow auf allen Text-Elementen
- **Animierter K8s-Badge** (nur Catalog): Rotation und Skalierung

📚 **Technische Details**: Siehe [ANIMATIONS.md](ANIMATIONS.md) für vollständige Dokumentation aller Animationseffekte.

## Design-Elemente

### Farben
- **Primär (Kubernetes-Blau)**: `#326CE5`
- **Sekundär (Dunkelblau)**: `#1A4D8F`
- **Akzent (Gold/Gelb)**: `#FFD700`, `#FFA500`
- **Routing-Pfade**: 
  - Grün: `#00FF7F` (Primäre Route/Table 100)
  - Rot: `#FF6B6B` (Sekundäre Route/Table 200)
- **Text**: `#FFFFFF` (Weiß), `#E0E0E0` (Hellgrau)

### Symbolik
- **Netzwerk-Knoten (Goldene Kreise)**: Repräsentieren Kubernetes-Nodes und Services
- **Verbindungslinien**: Symbolisieren Netzwerk-Verbindungen
- **Gestrichelte Ringe**: Repräsentieren Policy-Bereiche/CIDR-Ranges
- **Routing-Pfeile**: 
  - Grüner Pfeil → Routing-Tabelle 100
  - Roter Pfeil → Routing-Tabelle 200
  - Zeigen Policy-Based Routing

## Verwendung

### In Markdown (GitHub README)
```markdown
![IP Rule Operator](docs/logo.svg)
```

### In HTML
```html
<img src="docs/logo.svg" alt="IP Rule Operator" width="200"/>
```

### Für OLM/Operator Catalog

Die `logo-catalog.svg` Datei sollte im Bundle-Metadata referenziert werden:

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

#### Base64-Encoding für OLM

```bash
# Linux/Mac/WSL
base64 -w 0 docs/logo-catalog.svg

# Windows PowerShell
[Convert]::ToBase64String([IO.File]::ReadAllBytes("docs\logo-catalog.svg"))
```

## PNG-Export (Optional)

Falls PNG-Versionen benötigt werden:

```bash
# Mit Inkscape
inkscape logo.svg --export-type=png --export-filename=logo.png --export-width=512 --export-height=512

# Mit ImageMagick
convert -background none logo.svg -resize 512x512 logo.png

# Für Catalog (mit Hintergrund)
convert logo-catalog.svg -resize 256x256 logo-catalog.png
```

## Empfohlene Größen

| Verwendung | Größe | Format | Datei |
|------------|-------|--------|-------|
| GitHub README Header | 200x200px | SVG | logo.svg |
| OperatorHub Catalog | 256x256px | SVG/PNG | logo-catalog.svg |
| Dokumentation | 150-200px | SVG | logo.svg |
| Website/Blog (klein) | 64x64px | PNG | logo-64.png |
| Website/Blog (mittel) | 128x128px | PNG | logo-128.png |
| Website/Blog (groß) | 512x512px | PNG | logo-512.png |
| Favicon | 32x32px | PNG/ICO | favicon.ico |

## Lizenz

Die Logos sind Teil des IP Rule Operator Projekts und unterliegen der Apache 2.0 Lizenz.

Copyright 2025 Marius Bertram.
