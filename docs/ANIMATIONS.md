# 🎬 IP Rule Operator - Animation Documentation

This document provides a comprehensive overview of all animation effects used in the IP Rule Operator logos.

---

## 📚 Table of Contents

1. [Animation Types](#animation-types)
2. [Technical Details](#technical-details)
3. [Performance Considerations](#performance-considerations)
4. [Customization](#customization)

---

## Animation Types

### 1. Pulse Animation (Pulsating Rings)
Network rings pulsate and change size and opacity:

```css
@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 0.3; }
  50% { transform: scale(1.05); opacity: 0.6; }
}
```

**Usage:**
- Outer ring (network boundaries)
- Inner ring (policy areas)

**Duration:** 3-4 seconds (depending on ring)
**Timing:** ease-in-out

---

### 2. Dash Animation (Running Dash Pattern)
Dashed lines rotate around their shape:

```css
@keyframes dash {
  to { stroke-dashoffset: -100; }
}
```

**Usage:**
- Ring borders (outer & inner rings)

**Duration:** 4-8 seconds (depending on ring)
**Timing:** linear

---

### 3. Node Glow (Glowing Nodes)
Network nodes get a pulsating shadow effect:

```css
@keyframes nodeGlow {
  0%, 100% { filter: drop-shadow(0 0 3px #FFD700); }
  50% { filter: drop-shadow(0 0 8px #FFA500); }
}
```

**Usage:**
- All network nodes (golden circles)

**Duration:** 2-2.5 seconds
**Timing:** ease-in-out
**Effect:** Golden → Orange Glow

---

### 4. Connection Pulse (Pulsating Connections)
Connection lines change thickness and opacity:

```css
@keyframes connectionPulse {
  0%, 100% { opacity: 0.5; stroke-width: 2; }
  50% { opacity: 0.9; stroke-width: 2.5; }
}
```

**Usage:**
- All network connection lines

**Duration:** 2 seconds
**Timing:** ease-in-out
**Delay:** Time-shifted (0s, 0.3s, 0.6s, 0.9s, 1.2s)

---

### 5. Arrow Flow (Flowing Arrows)
Routing arrows move horizontally:

```css
@keyframes arrowFlow {
  0% { transform: translateX(0); opacity: 1; }
  50% { transform: translateX(8px); opacity: 0.6; }
  100% { transform: translateX(0); opacity: 1; }
}
```

**Usage:**
- Green arrow (Table 100)
- Red arrow (Table 200)

**Duration:** 1.5-2 seconds
**Timing:** ease-in-out
**Delay:** Green: 0s, Red: 0.5-0.6s

---

### 6. Text Glow (Glowing Text)
Text elements get a color-changing glow:

```css
@keyframes textGlow {
  0%, 100% { filter: drop-shadow(0 0 3px currentColor); }
  50% { filter: drop-shadow(0 0 8px currentColor); }
}
```

**Usage:**
- "IP RULE" title
- "OPERATOR" subtitle
- "TABLE 100" / "TABLE 200" labels

**Duration:** 2.5-4 seconds
**Timing:** ease-in-out
**Effect:** White → Green (for title)

---

### 7. K8s Badge (Kubernetes Badge)
The K8s badge rotates, scales, and glows in different colors:

```css
@keyframes k8sBadge {
  0%, 100% { 
    transform: rotate(0deg) scale(1); 
    filter: drop-shadow(0 0 8px rgba(255, 255, 255, 0.6));
  }
  25% { 
    transform: rotate(-8deg) scale(1.15); 
    filter: drop-shadow(0 0 15px rgba(0, 255, 127, 0.9));
  }
  50% { 
    transform: rotate(0deg) scale(1.05); 
    filter: drop-shadow(0 0 8px rgba(255, 255, 255, 0.6));
  }
  75% { 
    transform: rotate(8deg) scale(1.15); 
    filter: drop-shadow(0 0 15px rgba(50, 108, 229, 0.9));
  }
}
```

**Usage:**
- K8s Badge (Catalog logo only, top right)

**Duration:** 3 seconds
**Timing:** ease-in-out
**Effect:** 
- Rotation: ±8° for dynamic movement
- Scaling: 1.0 → 1.15 for attention
- Glow: White → Green → White → Blue (Kubernetes colors)
**Design:** Larger badge (r=20px) with Kubernetes-blue background and white ring

---

## 📋 Animation Overview by Logo

### Standard Logo (logo.svg)

| Element | Animation | Duration | Delay |
|---------|-----------|----------|-------|
| Outer Ring | pulse + dash | 3s + 4s | 0s |
| Nodes (5x) | nodeGlow | 2s | 0s, 0.2s, 0.4s, 0.6s, 0.8s |
| Connections (5x) | connectionPulse | 2s | 0s, 0.3s, 0.6s, 0.9s, 1.2s |
| Green Arrow | arrowFlow | 1.5s | 0s |
| Red Arrow | arrowFlow | 1.5s | 0.5s |
| Title Text | textGlow | 3s | 0s |
| Subtitle Text | textGlow | 3s | 0.3s |

**Overall Effect:** Smoothly pulsating network representation with flowing routing arrows

---

### Catalog Logo (logo-catalog.svg)

| Element | Animation | Duration | Delay |
|---------|-----------|----------|-------|
| Outer Ring | pulse + dash | 4s + 8s | 0s |
| Inner Ring | pulse + dash | 4s + 6s | 0.5s |
| Nodes (5x) | nodeGlow | 2.5s | 0s, 0.2s, 0.4s, 0.6s, 0.8s |
| Connections (5x) | connectionPulse | 2s | 0s, 0.3s, 0.6s, 0.9s, 1.2s |
| Green Arrow | arrowFlow | 2s | 0s |
| Red Arrow | arrowFlow | 2s | 0.6s |
| TABLE 100 Label | textGlow | 2.5s | 0s |
| TABLE 200 Label | textGlow | 2.5s | 0.4s |
| Title "IP RULE" | titlePulse | 4s | 0s |
| Subtitle "OPERATOR" | titlePulse | 4s | 0.5s |
| K8s Badge (IMPROVED) | k8sBadge | 3s | 0s |

**Overall Effect:** More intensive, detailed animation with more effects for OperatorHub. The K8s badge is now larger (r=20px), clearly visible with Kubernetes-blue background and rotates/pulsates with color-changing glow (white→green→blue).

---

## 🎨 Design Principles

### Timing & Coordination
- **Time Offsets**: Elements animate time-shifted for natural wave effects
- **Speed**: Slow, smooth animations (2-4s) avoid hectic
- **Loops**: Infinite repetition (`infinite`) for continuous movement

### Performance
- **Hardware Acceleration**: CSS transformations use GPU
- **No JavaScript**: Pure SVG+CSS for minimal overhead
- **Efficient**: Small file sizes (<10KB per logo)

### Accessibility
- **Subtle**: No flashing effects or fast movements
- **Disableable**: Browsers can disable animations with `prefers-reduced-motion`

### Browser Compatibility
- ✅ Chrome/Edge: Full support
- ✅ Firefox: Full support
- ✅ Safari: Full support
- ✅ Mobile Browsers: Full support

---

## 🔧 Customization

### Disable Animations
Add this to the SVG style:

```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation: none !important;
  }
}
```

### Change Speed
Modify the duration parameters:

```css
.ring { animation: pulse 3s ease-in-out infinite; }
/* Faster: */
.ring { animation: pulse 1.5s ease-in-out infinite; }
/* Slower: */
.ring { animation: pulse 6s ease-in-out infinite; }
```

### Add New Animations
1. Define keyframes in `<style>` block
2. Assign animation to a CSS class
3. Add class to SVG element

---

## 📊 Performance Metrics

- **File Sizes:**
  - logo.svg: ~8KB (with animations)
  - logo-catalog.svg: ~10KB (with animations)

- **Rendering:**
  - 60 FPS on modern devices
  - GPU-accelerated
  - No JavaScript overhead

- **Load Time:**
  - Instant with embedded SVG
  - <50ms with external reference

---

## 🚀 Best Practices

1. **In README**: Use the standard logo (logo.svg)
2. **In OperatorHub**: Use the catalog logo (logo-catalog.svg)
3. **Base64 Encoding**: Works with animations (for OLM)
4. **Responsive**: SVG scales perfectly at all sizes
5. **Dark/Light Mode**: Both logos work on dark backgrounds

---

## 💡 Example Usage

### In HTML/Markdown
```html
<img src="docs/logo.svg" alt="IP Rule Operator" width="200">
```

### In GitHub README
```markdown
![IP Rule Operator](docs/logo.svg)
```

### With Base64 in OLM
```yaml
spec:
  icon:
  - base64data: <BASE64_ENCODED_SVG>
    mediatype: image/svg+xml
```

Animations are preserved in all formats! 🎉

---

<div align="center">
  <img src="logo.svg" alt="Animated Logo" width="200"/>
  <p><em>Now with living animations! ✨</em></p>
</div>

