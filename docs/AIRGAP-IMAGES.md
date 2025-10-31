# IP Rule Operator - Image Liste für Air-Gapped Installation

## Übersicht
Diese Datei listet alle Container-Images auf, die für die Installation des IP Rule Operators in einem Air-Gapped Kubernetes/OpenShift-Cluster benötigt werden.

## Produktions-Images (Required für Air-Gapped Installation)

### 1. Controller Manager
```
ghcr.io/mariusbertram/iprule-controller:v0.0.1
```
- **Digest:** `ghcr.io/mariusbertram/iprule-controller@sha256:3b86a4a96a421aa75e46d0c9e6988dfba4cf1cf304309f65ec1637239cd104f2`
- **Beschreibung:** Hauptoperator-Controller für IPRule und IPRuleConfig CRDs
- **Deployment:** Kubernetes Deployment
- **Replicas:** 1
- **Privilegien:** Keine erhöhten Rechte erforderlich

### 2. Agent (DaemonSet)
```
ghcr.io/mariusbertram/iprule-agent:v0.0.1
```
- **Digest:** Muss bei Image-Push erstellt werden
- **Beschreibung:** Node-Agent für Anwendung der IP-Routing-Regeln
- **Deployment:** Kubernetes DaemonSet
- **Privilegien:** NET_ADMIN Capability, hostNetwork: true
- **Läuft auf:** Jedem Node im Cluster

---

## ClusterServiceVersion (CSV) - relatedImages Sektion

Die OLM ClusterServiceVersion enthält folgende Images in `spec.relatedImages`:

```yaml
spec:
  relatedImages:
    - image: ghcr.io/mariusbertram/iprule-controller@sha256:3b86a4a96a421aa75e46d0c9e6988dfba4cf1cf304309f65ec1637239cd104f2
      name: manager
    - image: ghcr.io/mariusbertram/iprule-agent@sha256:<DIGEST>
      name: agent
```

**Hinweis:** Die Images werden mit SHA256-Digests referenziert für garantierte Reproduzierbarkeit in Air-Gapped Umgebungen.

---

## Bundle Generation mit Image Digests

### Automatische Digest-Resolution

Das `make bundle` Kommando fügt automatisch beide Images mit Digests zur CSV hinzu:

```bash
make bundle \
  IMG=ghcr.io/mariusbertram/iprule-controller:v0.0.1 \
  AGENT_IMG=ghcr.io/mariusbertram/iprule-agent:v0.0.1 \
  VERSION=0.0.1 \
  USE_IMAGE_DIGESTS=true
```

**Was passiert:**
1. `operator-sdk generate bundle` erstellt die CSV mit Controller-Image (digest wird automatisch aufgelöst)
2. Script `hack/add-agent-to-csv.sh` fügt Agent-Image zu relatedImages hinzu
3. Wenn `USE_IMAGE_DIGESTS=true`: Script versucht Digest via skopeo/docker/podman aufzulösen
4. Bundle wird validiert

### Digest-Resolution Priorität

Das Script versucht folgende Tools in dieser Reihenfolge:

1. **skopeo** (empfohlen, funktioniert ohne Image pull)
   ```bash
   skopeo inspect docker://ghcr.io/mariusbertram/iprule-agent:v0.0.1
   ```

2. **docker inspect** (benötigt gepulltes Image)
   ```bash
   docker pull ghcr.io/mariusbertram/iprule-agent:v0.0.1
   docker inspect ghcr.io/mariusbertram/iprule-agent:v0.0.1
   ```

3. **podman inspect** (benötigt gepulltes Image)
   ```bash
   podman pull ghcr.io/mariusbertram/iprule-agent:v0.0.1
   podman inspect ghcr.io/mariusbertram/iprule-agent:v0.0.1
   ```

4. **Fallback:** Verwendet Tag wenn keine Digest-Resolution möglich

---

## Air-Gapped Installation - Vorbereitung

### Schritt 1: Images in Private Registry migrieren

```bash
# Source Registry (Internet-connected)
SOURCE_REGISTRY="ghcr.io/mariusbertram"

# Target Registry (Air-Gapped)
TARGET_REGISTRY="registry.airgap.local/ip-rule-operator"

# Version
VERSION="v0.0.1"

# Images mit skopeo kopieren (inklusive Digest)
skopeo copy \
  docker://${SOURCE_REGISTRY}/iprule-controller:${VERSION} \
  docker://${TARGET_REGISTRY}/iprule-controller:${VERSION}

skopeo copy \
  docker://${SOURCE_REGISTRY}/iprule-agent:${VERSION} \
  docker://${TARGET_REGISTRY}/iprule-agent:${VERSION}
```

### Schritt 2: Bundle für Air-Gapped Registry anpassen

```bash
# Bundle mit neuer Registry generieren
make bundle \
  IMG=${TARGET_REGISTRY}/iprule-controller:${VERSION} \
  AGENT_IMG=${TARGET_REGISTRY}/iprule-agent:${VERSION} \
  VERSION=0.0.1 \
  USE_IMAGE_DIGESTS=true
```

### Schritt 3: Bundle-Image erstellen und übertragen

```bash
# Bundle Image bauen
make bundle-build BUNDLE_IMG=${TARGET_REGISTRY}/bundle:${VERSION}

# Bundle Image pushen (in Air-Gapped Registry)
make bundle-push BUNDLE_IMG=${TARGET_REGISTRY}/bundle:${VERSION}
```

---

## OpenShift Air-Gapped Deployment

### ImageContentSourcePolicy (ICSP) erstellen

Für OpenShift sollten Sie eine ICSP erstellen, um Image-Mirrors zu konfigurieren:

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

Anwenden:
```bash
oc apply -f icsp.yaml
```

**Hinweis:** Nodes werden neu gestartet, um ICSP anzuwenden!

---

## Image-Liste für Mirroring (Skopeo Format)

Für automatisiertes Mirroring können Sie folgende Liste verwenden:

```yaml
# images.yaml
images:
  - source: ghcr.io/mariusbertram/iprule-controller:v0.0.1
    target: registry.airgap.local/ip-rule-operator/iprule-controller:v0.0.1
  - source: ghcr.io/mariusbertram/iprule-agent:v0.0.1
    target: registry.airgap.local/ip-rule-operator/iprule-agent:v0.0.1
```

Mirroring-Script:
```bash
#!/bin/bash
# mirror-images.sh

SOURCE_REGISTRY="ghcr.io/mariusbertram"
TARGET_REGISTRY="registry.airgap.local/ip-rule-operator"
VERSION="v0.0.1"

IMAGES=(
  "iprule-controller"
  "iprule-agent"
)

for image in "${IMAGES[@]}"; do
  echo "Mirroring ${image}..."
  skopeo copy \
    --all \
    docker://${SOURCE_REGISTRY}/${image}:${VERSION} \
    docker://${TARGET_REGISTRY}/${image}:${VERSION}
done

echo "All images mirrored successfully!"
```

---

## Digest-Verification

### Digests nach dem Push verifizieren

```bash
# Controller Digest abrufen
skopeo inspect docker://registry.airgap.local/ip-rule-operator/iprule-controller:v0.0.1 \
  | jq -r '.Digest'

# Agent Digest abrufen
skopeo inspect docker://registry.airgap.local/ip-rule-operator/iprule-agent:v0.0.1 \
  | jq -r '.Digest'
```

### CSV manuell mit Digests aktualisieren (falls erforderlich)

```bash
# Digest in CSV einfügen
CONTROLLER_DIGEST=$(skopeo inspect docker://registry.airgap.local/ip-rule-operator/iprule-controller:v0.0.1 | jq -r '.Digest')
AGENT_DIGEST=$(skopeo inspect docker://registry.airgap.local/ip-rule-operator/iprule-agent:v0.0.1 | jq -r '.Digest')

# CSV aktualisieren
sed -i "s|iprule-controller:v0.0.1|iprule-controller@${CONTROLLER_DIGEST}|g" bundle/manifests/ip-rule-operator.clusterserviceversion.yaml
sed -i "s|iprule-agent:v0.0.1|iprule-agent@${AGENT_DIGEST}|g" bundle/manifests/ip-rule-operator.clusterserviceversion.yaml
```

---

## Troubleshooting

### Problem: Digest kann nicht aufgelöst werden

**Symptom:**
```
⚠ Using tag instead of digest
```

**Lösungen:**

1. **skopeo installieren (empfohlen):**
   ```bash
   # RHEL/CentOS/Fedora
   sudo dnf install skopeo
   
   # Ubuntu/Debian
   sudo apt-get install skopeo
   ```

2. **Image vorab pullen:**
   ```bash
   docker pull ghcr.io/mariusbertram/iprule-agent:v0.0.1
   make bundle IMG=... AGENT_IMG=... USE_IMAGE_DIGESTS=true
   ```

3. **Digest manuell angeben:**
   ```bash
   AGENT_IMG="ghcr.io/mariusbertram/iprule-agent@sha256:..."
   make bundle IMG=... AGENT_IMG=$AGENT_IMG
   ```

### Problem: Image Pull in Air-Gapped Cluster schlägt fehl

**Lösungen:**

1. **ImagePullSecrets prüfen:**
   ```bash
   oc get secret -n ip-rule-operator-system
   ```

2. **ICSP Status prüfen:**
   ```bash
   oc get imagecontentsourcepolicy
   oc describe imagecontentsourcepolicy ip-rule-operator-mirrors
   ```

3. **Node-Status prüfen:**
   ```bash
   oc get nodes
   oc get mcp
   ```

---

## Zusammenfassung

### Minimal erforderliche Images für Betrieb:
1. **ghcr.io/mariusbertram/iprule-controller:v0.0.1** (mit Digest)
2. **ghcr.io/mariusbertram/iprule-agent:v0.0.1** (mit Digest)

### Für OLM-Installation zusätzlich:
3. **ghcr.io/mariusbertram/ip-rule-operator-bundle:v0.0.1**
4. **ghcr.io/mariusbertram/ip-rule-operator-catalog:v0.0.1** (optional)

### Automatisierung:
- `make bundle` mit `USE_IMAGE_DIGESTS=true` fügt beide Images automatisch mit Digests zur CSV hinzu
- Script `hack/add-agent-to-csv.sh` kümmert sich um Agent-Image-Eintrag
- Vollständige Unterstützung für Air-Gapped/Disconnected Deployments

---

**Letzte Aktualisierung:** 2025-10-31  
**Version:** 0.0.1  
**Zielplattform:** OpenShift 4.x (Air-Gapped)

