# Logo-Animationen - Technische Dokumentation

## 🎬 Übersicht

Die IP Rule Operator Logos verwenden moderne SVG+CSS Animationen, um eine dynamische und aufmerksamkeitsstarke Darstellung zu bieten. Alle Animationen laufen direkt im Browser ohne externe Abhängigkeiten.

## 📊 Animationstypen

### 1. Pulse Animation (Pulsieren)
Elemente werden größer/kleiner und ändern ihre Opacity:

```css
@keyframes pulse {
  0%, 100% { opacity: 0.4; transform: scale(1); }
  50% { opacity: 0.8; transform: scale(1.02); }
}
```

**Verwendung:**
- Netzwerkringe
- K8s Badge (mit Rotation)

**Dauer:** 3-4 Sekunden
**Timing:** ease-in-out

---

### 2. Dash Animation (Laufendes Strichmuster)
Gestrichelte Linien rotieren um ihre Form:

```css
@keyframes dash {
  to { stroke-dashoffset: -100; }
}
```

**Verwendung:**
- Ring-Grenzen (outer & inner rings)

**Dauer:** 4-8 Sekunden (abhängig vom Ring)
**Timing:** linear

---

### 3. Node Glow (Leuchtende Knoten)
Netzwerk-Nodes bekommen einen pulsierenden Schatten-Effekt:

```css
@keyframes nodeGlow {
  0%, 100% { filter: drop-shadow(0 0 3px #FFD700); }
  50% { filter: drop-shadow(0 0 8px #FFA500); }
}
```

**Verwendung:**
- Alle Netzwerk-Knoten (goldene Kreise)

**Dauer:** 2-2.5 Sekunden
**Timing:** ease-in-out
**Effekt:** Goldener → Oranger Glow

---

### 4. Connection Pulse (Pulsierende Verbindungen)
Verbindungslinien ändern Dicke und Opacity:

```css
@keyframes connectionPulse {
  0%, 100% { opacity: 0.5; stroke-width: 2; }
  50% { opacity: 0.9; stroke-width: 2.5; }
}
```

**Verwendung:**
- Alle Netzwerk-Verbindungslinien

**Dauer:** 2 Sekunden
**Timing:** ease-in-out
**Verzögerung:** Zeitversetzt (0s, 0.3s, 0.6s, 0.9s, 1.2s)

---

### 5. Arrow Flow (Fließende Pfeile)
Routing-Pfeile bewegen sich horizontal:

```css
@keyframes arrowFlow {
  0% { transform: translateX(0); opacity: 1; }
  50% { transform: translateX(8px); opacity: 0.6; }
  100% { transform: translateX(0); opacity: 1; }
}
```

**Verwendung:**
- Grüner Pfeil (Table 100)
- Roter Pfeil (Table 200)

**Dauer:** 1.5-2 Sekunden
**Timing:** ease-in-out
**Verzögerung:** Grün: 0s, Rot: 0.5-0.6s

---

### 6. Text Glow (Leuchtender Text)
Text-Elemente bekommen einen farbwechselnden Glow:

```css
@keyframes textGlow {
  0%, 100% { filter: drop-shadow(0 0 3px currentColor); }
  50% { filter: drop-shadow(0 0 8px currentColor); }
}
```

**Verwendung:**
- "IP RULE" Titel
- "OPERATOR" Untertitel
- "TABLE 100" / "TABLE 200" Labels

**Dauer:** 2.5-4 Sekunden
**Timing:** ease-in-out
**Effekt:** Weiß → Grün (beim Titel)

---

### 7. K8s Badge (Kubernetes Badge)
Der K8s-Badge rotiert, skaliert und leuchtet in verschiedenen Farben:

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

**Verwendung:**
- K8s Badge (nur Catalog-Logo, oben rechts)

**Dauer:** 3 Sekunden
**Timing:** ease-in-out
**Effekt:** 
- Rotation: ±8° für dynamische Bewegung
- Skalierung: 1.0 → 1.15 für Aufmerksamkeit
- Glow: Weiß → Grün → Weiß → Blau (Kubernetes-Farben)
**Design:** Größerer Badge (r=20px) mit Kubernetes-blauem Hintergrund und weißem Ring

---

## 📋 Animations-Übersicht nach Logo

### Standard-Logo (logo.svg)

| Element | Animation | Dauer | Verzögerung |
|---------|-----------|-------|-------------|
| Outer Ring | pulse + dash | 3s + 4s | 0s |
| Nodes (5x) | nodeGlow | 2s | 0s, 0.2s, 0.4s, 0.6s, 0.8s |
| Connections (5x) | connectionPulse | 2s | 0s, 0.3s, 0.6s, 0.9s, 1.2s |
| Green Arrow | arrowFlow | 1.5s | 0s |
| Red Arrow | arrowFlow | 1.5s | 0.5s |
| Title Text | textGlow | 3s | 0s |
| Subtitle Text | textGlow | 3s | 0.3s |

**Gesamt-Effekt:** Sanft pulsierende Netzwerk-Darstellung mit fließenden Routing-Pfeilen

---

### Catalog-Logo (logo-catalog.svg)

| Element | Animation | Dauer | Verzögerung |
|---------|-----------|-------|-------------|
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
| K8s Badge (VERBESSERT) | k8sBadge | 3s | 0s |

**Gesamt-Effekt:** Intensivere, detaillierte Animation mit mehr Effekten für OperatorHub. Der K8s-Badge ist jetzt größer (r=20px), deutlich sichtbar mit Kubernetes-blauem Hintergrund und rotiert/pulsiert mit farbwechselndem Glow (weiß→grün→blau).

---

## 🎯 Design-Prinzipien

### Timing & Koordination
- **Zeitversätze**: Elemente animieren zeitversetzt für natürliche Welleneffekte
- **Geschwindigkeit**: Langsame, sanfte Animationen (2-4s) vermeiden Hektik
- **Loops**: Endlose Wiederholung (`infinite`) für kontinuierliche Bewegung

### Performance
- **Hardware-Beschleunigung**: CSS-Transformationen nutzen GPU
- **Keine JavaScript**: Reine SVG+CSS für minimale Overhead
- **Effizient**: Kleine Dateigrößen (<10KB pro Logo)

### Barrierefreiheit
- **Subtil**: Keine blinkenden Effekte oder schnelle Bewegungen
- **Deaktivierbar**: Browser können Animationen mit `prefers-reduced-motion` deaktivieren

### Browser-Kompatibilität
- ✅ Chrome/Edge: Volle Unterstützung
- ✅ Firefox: Volle Unterstützung
- ✅ Safari: Volle Unterstützung
- ✅ Mobile Browser: Volle Unterstützung

---

## 🔧 Anpassungen

### Animation deaktivieren
Füge dies zum SVG-Style hinzu:

```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation: none !important;
  }
}
```

### Geschwindigkeit ändern
Ändere die Dauer-Parameter:

```css
.ring { animation: pulse 3s ease-in-out infinite; }
/* Schneller: */
.ring { animation: pulse 1.5s ease-in-out infinite; }
/* Langsamer: */
.ring { animation: pulse 6s ease-in-out infinite; }
```

### Neue Animationen hinzufügen
1. Definiere Keyframes im `<style>` Block
2. Weise Animation einer CSS-Klasse zu
3. Füge Klasse zum SVG-Element hinzu

---

## 📊 Performance-Metriken

- **Dateigrößen:**
  - logo.svg: ~8KB (mit Animationen)
  - logo-catalog.svg: ~10KB (mit Animationen)

- **Rendering:**
  - 60 FPS auf modernen Geräten
  - GPU-beschleunigt
  - Kein JavaScript-Overhead

- **Ladezeit:**
  - Instant bei eingebettetem SVG
  - <50ms bei externer Referenz

---

## 🚀 Best Practices

1. **Im README**: Verwende das Standard-Logo (logo.svg)
2. **In OperatorHub**: Verwende das Catalog-Logo (logo-catalog.svg)
3. **Base64-Encoding**: Funktioniert mit Animationen (für OLM)
4. **Responsive**: SVG skaliert perfekt auf allen Größen
5. **Dark/Light Mode**: Beide Logos funktionieren auf dunklem Hintergrund

---

## 📝 Beispiel-Verwendung

### In HTML/Markdown
```html
<img src="docs/logo.svg" alt="IP Rule Operator" width="200">
```

### In GitHub README
```markdown
![IP Rule Operator](docs/logo.svg)
```

### Mit Base64 in OLM
```yaml
spec:
  icon:
  - base64data: <BASE64_ENCODED_SVG>
    mediatype: image/svg+xml
```

Die Animationen bleiben in allen Formaten erhalten! 🎉

---

<div align="center">
  <img src="logo.svg" alt="Animated Logo" width="200"/>
  <p><em>Jetzt mit lebendigen Animationen! ✨</em></p>
</div>

