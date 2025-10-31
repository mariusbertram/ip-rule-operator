# Agent Build und Deployment

## Übersicht

Dieses Dokument beschreibt den Build- und Deployment-Prozess für den `ip-rule-operator-agent`.

## Neue Dateien

- **`Dockerfile.agent`** - Multi-Stage Docker Build für den Agent
- **`docs/RBAC-MANIFESTS-SOLUTION.md`** - Lösungen für RBAC und `make manifests`

## Neue Makefile Targets

### Build Targets

```bash
# Agent Binary lokal bauen (benötigt Linux Build-Tags)
make build-agent

# Beide Binaries bauen (Manager + Agent)
make build-all

# Agent lokal ausführen (benötigt Linux, NET_ADMIN Capabilities)
make run-agent
```

### Docker Targets

```bash
# Agent Docker Image bauen
make docker-build-agent

# Beide Images bauen (Manager + Agent)
make docker-build-all

# Agent Image pushen
make docker-push-agent

# Beide Images pushen
make docker-push-all

# Agent Multi-Platform Build (arm64, amd64, s390x, ppc64le)
make docker-buildx-agent
```

## Verwendung

### Lokaler Build (Linux)

```bash
# Agent Binary kompilieren
make build-agent

# Binary ausführen (benötigt CAP_NET_ADMIN)
sudo ./bin/agent
```

### Docker Build

```bash
# Standard Image bauen
make docker-build-agent AGENT_IMG=brtrm.dev/ip-rule-operator-agent:v0.0.1

# Image pushen
make docker-push-agent AGENT_IMG=brtrm.dev/ip-rule-operator-agent:v0.0.1

# Multi-Platform Build und Push
make docker-buildx-agent AGENT_IMG=brtrm.dev/ip-rule-operator-agent:v0.0.1
```

### Beide Images bauen

```bash
# Manager und Agent zusammen bauen
make docker-build-all \
  IMG=brtrm.dev/ip-rule-operator:v0.0.1 \
  AGENT_IMG=brtrm.dev/ip-rule-operator-agent:v0.0.1

# Beide pushen
make docker-push-all \
  IMG=brtrm.dev/ip-rule-operator:v0.0.1 \
  AGENT_IMG=brtrm.dev/ip-rule-operator-agent:v0.0.1
```

## Image Details

### Agent Image Eigenschaften

- **Base Image:** `gcr.io/distroless/static:nonroot`
- **Build Tags:** `linux` (für netlink Unterstützung)
- **CGO:** Disabled (statisches Binary)
- **User:** 65532:65532 (nonroot)
- **Plattformen:** linux/amd64, linux/arm64, linux/s390x, linux/ppc64le

### Benötigte Capabilities

Der Agent benötigt folgende Linux Capabilities zur Laufzeit:

- **`CAP_NET_ADMIN`** - Zum Verwalten von IP Rules
- **`hostNetwork: true`** - Zugriff auf Host-Netzwerk-Namespace

### Environment Variablen

```yaml
env:
- name: NODE_NAME
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
- name: RECONCILE_PERIOD
  value: "10s"  # Optional, default: 10s
```

## Deployment

### Als DaemonSet

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: ip-rule-agent
spec:
  selector:
    matchLabels:
      app: ip-rule-agent
  template:
    metadata:
      labels:
        app: ip-rule-agent
    spec:
      hostNetwork: true
      serviceAccountName: iprule-agent
      containers:
      - name: agent
        image: brtrm.dev/ip-rule-operator-agent:v0.0.1
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
          privileged: false
          runAsNonRoot: true
          runAsUser: 65532
```

## Entwicklung

### Lokales Testen (WSL)

```bash
# In WSL RedHatEnterpriseLinux-10.0
wsl -d RedHatEnterpriseLinux-10.0

# Agent kompilieren
make build-agent

# Mit sudo ausführen (für NET_ADMIN)
sudo ./bin/agent
```

### Debug Modus

```go
// In cmd/agent/main.go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

## CI/CD Integration

### GitHub Actions Beispiel

```yaml
name: Build Agent

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3
    
    - name: Login to Registry
      uses: docker/login-action@v3
      with:
        registry: brtrm.dev
        username: ${{ secrets.REGISTRY_USERNAME }}
        password: ${{ secrets.REGISTRY_PASSWORD }}
    
    - name: Build and Push
      run: |
        make docker-buildx-agent \
          AGENT_IMG=brtrm.dev/ip-rule-operator-agent:${GITHUB_REF_NAME}
```

## Troubleshooting

### Build Fehler: "netlink not supported"

**Problem:** Build schlägt fehl auf nicht-Linux Systemen.

**Lösung:** Der Agent kann nur auf Linux gebaut werden. Nutzen Sie:
- WSL auf Windows
- Docker Build (Multi-Stage Build läuft auf Linux)
- CI/CD Pipeline

### Runtime Error: "operation not permitted"

**Problem:** Agent kann keine IP Rules erstellen.

**Lösung:** Stellen Sie sicher, dass:
- `hostNetwork: true` gesetzt ist
- `CAP_NET_ADMIN` Capability vorhanden ist
- Security Context Constraints (OpenShift) korrekt sind

### Koordiniertes Löschen funktioniert nicht

**Problem:** IPRuleConfigs mit `state: absent` werden nicht gelöscht.

**Lösung:** 
- `NODE_NAME` Environment Variable muss gesetzt sein
- ServiceAccount braucht Zugriff auf Node-Liste
- Alle Nodes müssen den Agent laufen haben

## Weitere Informationen

- [RBAC Manifests Lösung](./RBAC-MANIFESTS-SOLUTION.md)
- [Agent Hauptcode](../cmd/agent/main.go)
- [Kubernetes Operator Patterns](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

