# E2E Tests für IP-Rule-Operator

Diese Dokumentation beschreibt die End-to-End (E2E) Tests für den IP-Rule-Operator.

## Übersicht

Die E2E-Tests validieren die vollständige Funktionalität des Operators in einer echten Kubernetes-Umgebung. Sie verwenden:
- **Kind** (Kubernetes in Docker) für einen temporären Test-Cluster
- **Docker** für das Bauen und Laden von Images
- **Ginkgo** als Test-Framework
- **kubectl** für Kubernetes-Interaktionen

## Test-Struktur

### Dateien

1. **`e2e_suite_test.go`**: Test-Suite Setup
   - Baut das Operator-Image
   - Lädt das Image in den Kind-Cluster
   - Installiert CertManager (falls benötigt)

2. **`e2e_test.go`**: Haupt-Testfälle
   - Manager-Deployment Tests
   - Metrics-Endpoint Tests
   - IPRule CR Tests
   - Agent CR Tests

3. **`../utils/utils.go`**: Helper-Funktionen
   - kubectl-Wrapper
   - CertManager-Installation
   - Image-Loading

## Test-Szenarien

### 1. Manager Tests
- ✅ Verifiziert, dass der Controller-Manager Pod läuft
- ✅ Prüft Pod-Status (Running)
- ✅ Validiert korrekte Labels

### 2. Metrics Tests
- ✅ Erstellt ClusterRoleBinding für Metrics-Zugriff
- ✅ Verifiziert Metrics-Service Verfügbarkeit
- ✅ Testet Metrics-Endpoint über curl
- ✅ Prüft auf `controller_runtime_reconcile_total` Metrik

### 3. IPRule Tests
- ✅ Erstellt eine Test-IPRule CR
- ✅ Verifiziert erfolgreiche Erstellung
- ✅ Prüft, dass die Ressource im Cluster existiert
- ✅ Cleanup nach dem Test

### 4. Agent Tests
- ✅ Erstellt eine Test-Agent CR
- ✅ Verifiziert, dass ein DaemonSet erstellt wird
- ✅ Prüft Agent-Status Updates
- ✅ Validiert Condition-Felder
- ✅ Cleanup nach dem Test

## Ausführung

### Lokal mit Make

```bash
# Vollständige E2E-Tests (erstellt und zerstört Kind-Cluster)
make test-e2e

# Mit Docker statt Podman
make test-e2e CONTAINER_TOOL=docker

# Cluster-Setup ohne Tests
make setup-test-e2e

# Nur Cleanup
make cleanup-test-e2e
```

### Manuell

```bash
# 1. Kind-Cluster erstellen
kind create cluster --name ip-rule-operator-test-e2e

# 2. Image bauen
make docker-build IMG=example.com/ip-rule-operator:v0.0.1

# 3. Image in Kind laden
kind load docker-image example.com/ip-rule-operator:v0.0.1 --name ip-rule-operator-test-e2e

# 4. Tests ausführen
cd test/e2e
KIND_CLUSTER=ip-rule-operator-test-e2e go test -v -ginkgo.v

# 5. Cleanup
kind delete cluster --name ip-rule-operator-test-e2e
```

### In GitHub Actions

Die GitHub Action `.github/workflows/test-e2e.yml` führt die Tests automatisch bei jedem Push und Pull Request aus:

```yaml
- uses: actions/checkout@v4
- uses: actions/setup-go@v5
- Install Kind
- Run: make test-e2e CONTAINER_TOOL=docker
```

## Umgebungsvariablen

- `KIND_CLUSTER`: Name des Kind-Clusters (Standard: `ip-rule-operator-test-e2e`)
- `CERT_MANAGER_INSTALL_SKIP=true`: Überspringt CertManager-Installation
- `IMG`: Container-Image für den Operator (Standard: `example.com/ip-rule-operator:v0.0.1`)

## Voraussetzungen

### Lokal
- Docker oder Podman
- Kind CLI
- kubectl
- Go 1.24+
- Make

### GitHub Actions
- Wird automatisch bereitgestellt
- Ubuntu-latest Runner
- Docker ist vorinstalliert

## Debugging

### Test-Logs anzeigen

```bash
# Controller-Manager Logs
kubectl logs -n ip-rule-operator-system -l control-plane=controller-manager

# Events anzeigen
kubectl get events -n ip-rule-operator-system --sort-by=.lastTimestamp

# Pod-Status
kubectl describe pod -n ip-rule-operator-system -l control-plane=controller-manager
```

### Bei Test-Fehlern

Die E2E-Tests sammeln automatisch Debug-Informationen:
- Controller-Manager Logs
- Kubernetes Events
- Pod Descriptions
- Metrics-Curl Logs

Diese werden in GinkgoWriter ausgegeben, wenn ein Test fehlschlägt.

## Bekannte Limitierungen

1. **Netzwerk**: Tests laufen in einem isolierten Kind-Cluster
2. **LoadBalancer**: Keine echten LB-IPs in Kind (benötigt metallb oder ähnliches)
3. **Agent**: Agent-Pods benötigen NET_ADMIN Capability (funktioniert nicht in restriktiven Umgebungen)
4. **Timeouts**: Längere Timeouts für Image-Pulls und Pod-Starts

## Erweitern der Tests

Um neue E2E-Tests hinzuzufügen:

1. Füge neue `It()` Blocks in `e2e_test.go` hinzu
2. Verwende `utils.Run()` für kubectl-Befehle
3. Nutze `Eventually()` für asynchrone Validierungen
4. Stelle sicher, dass Cleanup in `AfterEach` oder im Test selbst erfolgt

Beispiel:

```go
It("should test my feature", func() {
    By("creating a test resource")
    // ... test code ...
    
    By("verifying expected behavior")
    verifyFunc := func(g Gomega) {
        // ... assertions ...
    }
    Eventually(verifyFunc, 30*time.Second).Should(Succeed())
    
    By("cleaning up")
    // ... cleanup code ...
})
```

## Troubleshooting

### "Kind cluster already exists"
```bash
kind delete cluster --name ip-rule-operator-test-e2e
```

### "Image not found in Kind"
```bash
make docker-build IMG=example.com/ip-rule-operator:v0.0.1
kind load docker-image example.com/ip-rule-operator:v0.0.1 --name ip-rule-operator-test-e2e
```

### "CRDs not installed"
```bash
kubectl get crds | grep iprule
make install
```

### Tests timeout
- Erhöhe die Timeouts in den Tests
- Prüfe Docker/Kind Ressourcen (CPU, Memory)
- Prüfe Image-Pull-Status

## Erfolgsmetriken

Ein erfolgreicher E2E-Test-Lauf sollte:
- ✅ Den Kind-Cluster erstellen
- ✅ Das Operator-Image bauen und laden
- ✅ CRDs installieren
- ✅ Den Operator deployen
- ✅ Alle Test-Cases bestehen (Manager, Metrics, IPRule, Agent)
- ✅ Den Cluster sauber aufräumen

Typische Laufzeit: **5-10 Minuten**

