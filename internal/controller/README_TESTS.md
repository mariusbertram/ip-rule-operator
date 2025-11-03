# Controller Tests

Diese Datei beschreibt die Test-Struktur für die Controller des IP-Rule-Operators.

## Test-Dateien

### 1. `suite_test.go`
- Setup und Teardown für die gesamte Test-Suite
- Initialisiert die envtest-Umgebung (Kubernetes API Server für Tests)
- Konfiguriert Ginkgo/Gomega als Test-Framework

### 2. `iprule_controller_test.go`
Umfassende Tests für den IPRule Controller:

#### Test Cases:
- **Basis-Reconcile**: Testet das Reconciling ohne Services
  - Stellt sicher, dass der Controller ohne Fehler läuft
  - Verifiziert, dass keine IPRuleConfigs erstellt werden, wenn keine Services vorhanden sind

- **LoadBalancer Service Integration**: Testet die Erstellung von IPRuleConfigs
  - Erstellt ein IPRule CRD mit CIDR 10.0.0.0/24
  - Erstellt einen LoadBalancer Service mit passender IP
  - Verifiziert, dass IPRuleConfig automatisch erstellt wird
  - Prüft korrekte Table- und Priority-Werte

### 3. `agent_controller_test.go`
Umfassende Tests für den Agent Controller:

#### Test Cases:
- **DaemonSet Erstellung**: Testet die Basis-Funktionalität
  - Erstellt eine Agent CR
  - Verifiziert, dass ein DaemonSet "iprule-agent" erstellt wird
  - Prüft korrekte Konfiguration (Image, NodeSelector, HostNetwork)

- **Status Synchronisation**: Testet Status-Updates
  - Verifiziert, dass Agent.Status mit DaemonSet-Status synchronisiert wird
  - Prüft, dass Conditions gesetzt werden

- **Multiple Agents**: Testet das Verhalten bei mehreren Agent-Instanzen
  - Erstellt zwei Agent CRs
  - Verifiziert, dass nur die alphabetisch erste aktiv ist
  - Prüft, dass die zweite als "InactiveInstance" markiert wird

### 4. `basic_test.go`
Unit Tests für Helper-Funktionen (keine Kubernetes-Integration):

#### Test Cases:
- **TestBuildDesiredEntryMap**: Testet die IP-Regel-Logik
  - Verifiziert korrekte Zuordnung von Service IPs zu Routing Tables
  - Testet die "most specific CIDR wins" Logik

- **TestComputeTemplateHash**: Testet Hash-Berechnung
  - Stellt sicher, dass identische Konfigurationen denselben Hash erzeugen
  - Verifiziert, dass unterschiedliche Configs unterschiedliche Hashes haben

- **TestUpsertCondition**: Testet Condition-Management
  - Testet das Hinzufügen neuer Conditions
  - Testet das Aktualisieren bestehender Conditions

- **TestFindCondition**: Testet Condition-Suche
  - Verifiziert das Finden existierender Conditions
  - Testet korrektes Verhalten bei nicht-existierenden Conditions

## Tests ausführen

### Alle Tests
```bash
wsl -d RedHatEnterpriseLinux-10.0 bash -c "cd /mnt/c/Users/mariu/GolandProjects/ip-rule-operator && make test"
```

### Nur Unit Tests (ohne envtest)
```bash
wsl -d RedHatEnterpriseLinux-10.0 bash -c "cd /mnt/c/Users/mariu/GolandProjects/ip-rule-operator && go test -v ./internal/controller/... -run '^Test[^C]'"
```

### Nur Controller Tests (mit envtest)
```bash
wsl -d RedHatEnterpriseLinux-10.0 bash -c "cd /mnt/c/Users/mariu/GolandProjects/ip-rule-operator && go test -v ./internal/controller/... -run TestControllers"
```

### Einzelne Tests
```bash
# IPRule Tests
wsl -d RedHatEnterpriseLinux-10.0 bash -c "cd /mnt/c/Users/mariu/GolandProjects/ip-rule-operator && go test -v ./internal/controller/... -run 'TestControllers/IpRule'"

# Agent Tests
wsl -d RedHatEnterpriseLinux-10.0 bash -c "cd /mnt/c/Users/mariu/GolandProjects/ip-rule-operator && go test -v ./internal/controller/... -run 'TestControllers/Agent'"

# Unit Tests
wsl -d RedHatEnterpriseLinux-10.0 bash -c "cd /mnt/c/Users/mariu/GolandProjects/ip-rule-operator && go test -v ./internal/controller/... -run TestBuildDesiredEntryMap"
```

## Voraussetzungen

- Go 1.24+
- WSL mit RedHatEnterpriseLinux-10.0
- envtest Binaries (werden automatisch von `make test` installiert)

## Test-Coverage

Die Tests decken folgende Bereiche ab:

1. **Controller-Logik**: Reconcile-Loops und Event-Handling
2. **CRD-Management**: Erstellung und Update von Custom Resources
3. **Service-Integration**: Überwachung von LoadBalancer Services
4. **Status-Management**: Synchronisation von Status-Feldern
5. **Helper-Funktionen**: Utility-Funktionen und Algorithmen

## Bekannte Einschränkungen

- Die envtest-Umgebung simuliert nur die Kubernetes API, keine echten Nodes
- DaemonSet Pods werden nicht wirklich gestartet (nur API-Objekte)
- LoadBalancer IP-Zuweisung muss manuell simuliert werden
- Netzwerk-Operations (ip rule add/del) werden nicht getestet

## Erweitern der Tests

Um neue Tests hinzuzufügen:

1. Für Controller-Tests: Füge neue `It()` Blocks in `*_controller_test.go` hinzu
2. Für Unit-Tests: Füge neue `func Test*()` Funktionen in `basic_test.go` hinzu
3. Stelle sicher, dass alle Tests idempotent sind (cleanup in AfterEach)
4. Verwende Eventually() für asynchrone Assertions

