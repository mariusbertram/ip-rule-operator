# IP Rule Operator - Logo-Vorschau

> ✨ **NEU:** Die Logos sind jetzt vollständig animiert für mehr Aufmerksamkeit!

## Standard-Logo (README)

<div align="center">
  <img src="logo.svg" alt="Standard Logo" width="200"/>
  <p><strong>logo.svg</strong> - 200x200px - 🎬 ANIMIERT</p>
  <p>Verwendung: GitHub README, Dokumentation</p>
  <p><em>Features: Pulsierende Ringe, leuchtende Nodes, fließende Pfeile</em></p>
</div>

---

## Catalog-Logo (OperatorHub)

<div align="center">
  <img src="logo-catalog.svg" alt="Catalog Logo" width="256"/>
  <p><strong>logo-catalog.svg</strong> - 256x256px - 🎬 ANIMIERT</p>
  <p>Verwendung: OpenShift OperatorHub, Operator Catalogs</p>
  <p><em>Features: Doppelt-pulsierende Ringe, animierter K8s-Badge, leuchtende Labels</em></p>
</div>

---

## Vergleich

| Feature | Standard-Logo | Catalog-Logo |
|---------|---------------|--------------|
| Größe | 200x200px | 256x256px |
| Hintergrund | Transparent/Blau | Gradient mit Rounded Corners |
| Effekte | ✨ Animationen | Drop-Shadow + ✨ Animationen |
| Details | Einfach | Detailliert mit Labels |
| Verwendung | README, Docs | OperatorHub, Catalogs |
| **Animationen** | **7 Effekte** | **11 Effekte** |

---

## 🎬 Animations-Features

### Standard-Logo (logo.svg)
- ✅ Pulsierende Netzwerkringe (3s Zyklen)
- ✅ Leuchtende Nodes mit Glow-Effekt (2s Zyklen)
- ✅ Pulsierende Verbindungslinien (zeitversetzt)
- ✅ Fließende Routing-Pfeile (grün & rot)
- ✅ Leuchtender Text mit dynamischem Glow

### Catalog-Logo (logo-catalog.svg)
- ✅ Doppelte pulsierende Ringe (outer & inner)
- ✅ Intensive Node-Glow-Effekte
- ✅ Pulsierende Verbindungen mit Dickenänderung
- ✅ Fließende Routing-Pfeile mit Opacity-Änderung
- ✅ Leuchtende "TABLE 100" und "TABLE 200" Labels
- ✅ Animierter K8s-Badge (Rotation & Skalierung)
- ✅ Pulsierender Titel mit Farbwechsel

**Alle Animationen:** Reine SVG+CSS, keine JavaScript-Abhängigkeiten!

📚 **Mehr Details:** Siehe [ANIMATIONS.md](ANIMATIONS.md) für technische Dokumentation

---

## Design-Elemente

### Farb palette

- 🔵 **Kubernetes-Blau**: `#326CE5` - Primärfarbe
- 🔷 **Dunkelblau**: `#1A4D8F` - Akzente
- 🟡 **Gold**: `#FFD700` - Network Nodes
- 🟠 **Orange**: `#FFA500` - Node-Borders
- 🟢 **Grün**: `#00FF7F` - Routing Table 100
- 🔴 **Rot**: `#FF6B6B` - Routing Table 200
- ⚪ **Weiß**: `#FFFFFF` - Text & Lines
- ◻️ **Hellgrau**: `#E0E0E0` - Sekundärtext

### Symbole

- 🔵 **Großer Kreis**: Kubernetes-Cluster
- 🟡 **Goldene Knoten**: Kubernetes-Nodes/Services
- ⚪ **Weiße Linien**: Netzwerk-Verbindungen
- ⭕ **Gestrichelte Ringe**: Policy-Bereiche (CIDR)
- ➡️ **Grüner Pfeil**: Primäre Route (Table 100)
- ➡️ **Roter Pfeil**: Sekundäre Route (Table 200)
- 📝 **Text**: "IP RULE" & "OPERATOR"

---

## Quick Start

### Logo in README einbinden
```markdown
![IP Rule Operator](docs/logo.svg)
```

### Logo für OLM vorbereiten
```bash
# Linux/Mac/WSL
cd docs
./encode-logo.sh

# Windows PowerShell
cd docs
.\encode-logo.ps1
```

Die Base64-encodierte Version dann in die ClusterServiceVersion einfügen.

---

<div align="center">
  <p><em>Erstellt für das IP Rule Operator Projekt</em></p>
  <p>Apache 2.0 License | Copyright 2025 Marius Bertram</p>
</div>
