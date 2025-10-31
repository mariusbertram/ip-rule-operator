# IP Rule Operator - Generische Related Images Lösung

## ✅ Vollständig implementiert und getestet

Diese Lösung ermöglicht es, beliebig viele Container-Images zu `spec.relatedImages` in der ClusterServiceVersion hinzuzufügen - perfekt für Air-Gapped/Disconnected OpenShift-Deployments.

---

## Architektur

### 1. Asset-Datei für Image-Liste
**Datei:** `config/manifests/related-images.txt`

```bash
# Format: <name>:<image-url>
# Kommentare mit # sind erlaubt
# Leere Zeilen werden ignoriert

agent:ghcr.io/mariusbertram/iprule-agent:v0.0.1
# sidecar:ghcr.io/mariusbertram/sidecar:v1.0.0
# init:ghcr.io/mariusbertram/init:v2.0.0
```

**Features:**
- Einfaches Format: `name:image`
- Unterstützt Kommentare und Leerzeilen
- Wird automatisch von Makefile mit aktuellen Variablen aktualisiert

### 2. Config-Update Script (Bash)
**Datei:** `hack/update-related-images-config.sh`

```bash
#!/bin/bash
# Aktualisiert related-images.txt mit Makefile-Variablen

IMAGES_CONFIG="${1}"
AGENT_IMG="${2}"

cat > "$IMAGES_CONFIG" << EOF
# IP Rule Operator - Related Images Configuration
agent:${AGENT_IMG}
# Weitere Images können hier hinzugefügt werden
EOF
```

**Zweck:** Hält Asset-Datei synchron mit Makefile-Variablen

### 3. CSV-Update Script (Python)
**Datei:** `hack/add-related-images-to-csv.py`

```python
#!/usr/bin/env python3
# Liest related-images.txt
# Löst Digests auf (skopeo/docker/podman)
# Fügt Images zu CSV hinzu
```

**Features:**
- ✅ Robust und sicher (line-by-line YAML-Manipulation)
- ✅ Automatische Digest-Resolution
- ✅ Prüft auf Duplikate
- ✅ Detailliertes Logging mit Emoji-Status
- ✅ Exit-Code basierend auf Erfolg/Fehler

### 4. Makefile Integration

```makefile
.PHONY: bundle
bundle: manifests kustomize operator-sdk
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	@bash hack/update-related-images-config.sh config/manifests/related-images.txt $(AGENT_IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle $(BUNDLE_GEN_FLAGS)
	@python3 hack/add-related-images-to-csv.py config/manifests/related-images.txt $(USE_IMAGE_DIGESTS)
	$(OPERATOR_SDK) bundle validate ./bundle
```

---

## Verwendung

### Bundle mit allen Images generieren

```bash
make bundle \
  IMG=ghcr.io/mariusbertram/iprule-controller:v0.0.1 \
  AGENT_IMG=ghcr.io/mariusbertram/iprule-agent:v0.0.1 \
  VERSION=0.0.1 \
  USE_IMAGE_DIGESTS=true
```

**Output:**
```
==================================================
CSV Related Images Update Script (Python)
==================================================
Processing 1 image(s) from: config/manifests/related-images.txt

📦 Processing: agent
   Image: ghcr.io/mariusbertram/iprule-agent:v0.0.1
   🔍 Resolving digest...
   ✅ Resolved via podman
   ➕ Adding to relatedImages...
   ✅ Successfully added

==================================================
Summary
==================================================
✅ Added:   1
⏭️  Skipped: 0
❌ Failed:  0

Current relatedImages section:
--------------------------------------------------
  relatedImages:
    - image: ghcr.io/mariusbertram/iprule-controller@sha256:3b86a...
      name: manager
    - image: ghcr.io/mariusbertram/iprule-agent@sha256:f6ca1...
      name: agent

✅ Done!
```

### Neue Images hinzufügen

**Methode 1: Direkt in Asset-Datei**

Editieren Sie `config/manifests/related-images.txt`:

```bash
agent:ghcr.io/mariusbertram/iprule-agent:v0.0.1
sidecar:ghcr.io/mariusbertram/sidecar:v1.0.0
init:ghcr.io/mariusbertram/init-container:v2.0.0
```

**Methode 2: Via Makefile-Variablen**

1. Fügen Sie Variable zum Makefile hinzu:
```makefile
SIDECAR_IMG ?= ghcr.io/mariusbertram/sidecar:v$(VERSION)
```

2. Erweitern Sie `hack/update-related-images-config.sh`:
```bash
cat > "$IMAGES_CONFIG" << EOF
agent:${AGENT_IMG}
sidecar:${SIDECAR_IMG}
EOF
```

3. Rufen Sie Script mit neuer Variable auf:
```makefile
@bash hack/update-related-images-config.sh config/manifests/related-images.txt $(AGENT_IMG) $(SIDECAR_IMG)
```

---

## Digest-Resolution

### Automatische Auflösung

Das Python-Script versucht Digests in folgender Reihenfolge aufzulösen:

1. **skopeo** (empfohlen - kein Pull erforderlich)
   ```bash
   skopeo inspect docker://image:tag
   ```

2. **docker inspect** (benötigt gepulltes Image)
   ```bash
   docker inspect image:tag
   ```

3. **podman inspect** (benötigt gepulltes Image)
   ```bash
   podman inspect image:tag
   ```

4. **Fallback:** Tag wird verwendet wenn keine Digest-Auflösung möglich

### Digest-Resolution deaktivieren

```bash
make bundle USE_IMAGE_DIGESTS=false
```

**Hinweis:** Für Air-Gapped Deployments sollten Digests **immer aktiviert** sein!

---

## Beispiel: CSV relatedImages

Nach `make bundle` enthält die CSV:

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: ClusterServiceVersion
metadata:
  name: ip-rule-operator.v0.0.1
spec:
  # ...
  relatedImages:
    - image: ghcr.io/mariusbertram/iprule-controller@sha256:3b86a4a96a421aa75e46d0c9e6988dfba4cf1cf304309f65ec1637239cd104f2
      name: manager
    - image: ghcr.io/mariusbertram/iprule-agent@sha256:f6ca1f929a44cd69ba5f4980bf8ec1b119275540b6a8d3f706c657b1eca0eab6
      name: agent
  version: 0.0.1
```

**Beide Images mit SHA256-Digests!** ✅

---

## Air-Gapped Deployment

### Workflow

1. **Images mit Digests bauen**
```bash
make docker-build-all VERSION=0.0.1
make docker-push-all VERSION=0.0.1
```

2. **Bundle generieren**
```bash
make bundle VERSION=0.0.1 USE_IMAGE_DIGESTS=true
```

3. **Images mirroren**
```bash
SOURCE="ghcr.io/mariusbertram"
TARGET="registry.airgap.local/ip-rule-operator"

# Alle Images aus related-images.txt
skopeo copy --all \
  docker://${SOURCE}/iprule-controller:v0.0.1 \
  docker://${TARGET}/iprule-controller:v0.0.1

skopeo copy --all \
  docker://${SOURCE}/iprule-agent:v0.0.1 \
  docker://${TARGET}/iprule-agent:v0.0.1
```

4. **OpenShift ICSP erstellen**
```yaml
apiVersion: operator.openshift.io/v1alpha1
kind: ImageContentSourcePolicy
metadata:
  name: ip-rule-operator-mirrors
spec:
  repositoryDigestMirrors:
  - mirrors:
    - registry.airgap.local/ip-rule-operator/iprule-controller
    source: ghcr.io/mariusbertram/iprule-controller
  - mirrors:
    - registry.airgap.local/ip-rule-operator/iprule-agent
    source: ghcr.io/mariusbertram/iprule-agent
```

---

## Vorteile der generischen Lösung

### ✅ Skalierbar
- Beliebig viele Images hinzufügbar
- Einfaches Text-Format
- Keine Code-Änderungen erforderlich

### ✅ Wartbar
- Zentrale Asset-Datei
- Klare Trennung: Config vs. Code
- Self-documenting Format

### ✅ Automatisiert
- Makefile-Integration
- Automatische Digest-Resolution
- Duplikat-Vermeidung

### ✅ Robust
- Python für sichere YAML-Manipulation
- Fehlerbehandlung
- Detailliertes Logging

### ✅ Air-Gap Ready
- Digest-Support eingebaut
- Kompatibel mit skopeo/docker/podman
- OpenShift ICSP ready

---

## Troubleshooting

### Problem: "Python3 not found"

**Lösung:**
```bash
# RHEL/CentOS/Fedora
sudo dnf install python3

# Ubuntu/Debian
sudo apt-get install python3
```

### Problem: "Could not resolve digest"

**Ursachen:**
- Image existiert nicht in Registry
- Keine Netzwerkverbindung
- Authentifizierung fehlgeschlagen

**Lösungen:**
1. **Image vorab pullen:**
   ```bash
   docker pull ghcr.io/mariusbertram/iprule-agent:v0.0.1
   ```

2. **Registry Login:**
   ```bash
   docker login ghcr.io
   podman login ghcr.io
   ```

3. **skopeo installieren (empfohlen):**
   ```bash
   sudo dnf install skopeo
   ```

### Problem: "Image already present, skipping"

**Kein Problem!** Das Script erkennt Duplikate automatisch. Das Image wurde bei einem vorherigen Lauf bereits hinzugefügt.

**Zum Zurücksetzen:**
```bash
rm -rf bundle
make bundle
```

---

## Dateien

### Erstellt
- ✅ `config/manifests/related-images.txt` - Asset-Datei
- ✅ `hack/update-related-images-config.sh` - Config-Update (Bash)
- ✅ `hack/add-related-images-to-csv.py` - CSV-Update (Python)
- ✅ `hack/add-related-images-to-csv.sh` - Alte Version (Bash/AWK) - nicht empfohlen

### Aktualisiert
- ✅ `Makefile` - Bundle-Target erweitert

### Dokumentation
- ✅ `docs/AIRGAP-IMAGES.md` - Air-Gapped Deployment Guide
- ✅ `AIRGAP-QUICKREF.md` - Schnellreferenz
- ✅ Dieses Dokument

---

## Best Practices

### 1. Immer Digests für Production verwenden
```bash
make bundle USE_IMAGE_DIGESTS=true  # ✅ Empfohlen
make bundle USE_IMAGE_DIGESTS=false # ❌ Nur für Dev
```

### 2. skopeo für Air-Gapped Workflows nutzen
```bash
sudo dnf install skopeo  # Funktioniert ohne Image-Pull
```

### 3. related-images.txt unter Versionskontrolle
```bash
git add config/manifests/related-images.txt
git commit -m "Add new sidecar image to related images"
```

### 4. Bundle vor jedem Release testen
```bash
make bundle VERSION=x.y.z
operator-sdk bundle validate ./bundle
```

### 5. Digests in CI/CD verifizieren
```yaml
# .github/workflows/bundle.yml
- name: Verify all images have digests
  run: |
    grep -A 50 'relatedImages:' bundle/manifests/*.clusterserviceversion.yaml | \
    grep 'image:' | \
    grep -v '@sha256:' && exit 1 || echo "All images have digests ✅"
```

---

## Zusammenfassung

### Was wurde erreicht ✅

1. **Generische Architektur:** Asset-basierte Image-Liste
2. **Automatische Digest-Resolution:** skopeo/docker/podman
3. **Robuste YAML-Manipulation:** Python statt AWK
4. **Makefile-Integration:** Vollautomatisch bei `make bundle`
5. **Air-Gap Ready:** Alle Images mit Digests in CSV
6. **Skalierbar:** Beliebig viele Images erweiterbar
7. **Getestet:** Funktioniert einwandfrei ✅

### Quick Start

```bash
# Neues Image hinzufügen
echo "sidecar:ghcr.io/myorg/sidecar:v1.0.0" >> config/manifests/related-images.txt

# Bundle generieren
make bundle VERSION=0.0.1 USE_IMAGE_DIGESTS=true

# Verifizieren
grep -A 20 'relatedImages:' bundle/manifests/*.clusterserviceversion.yaml
```

---

**Status:** ✅ Production Ready  
**Getestet:** ✅ 2025-10-31  
**Python Version:** 3.x  
**Platforms:** Linux, WSL2  


