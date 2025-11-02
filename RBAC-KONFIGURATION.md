# RBAC-Konfiguration für IP-Rule-Operator (Korrekte Vorgehensweise)

## Übersicht
Die RBAC-Berechtigungen für den IP-Rule-Operator müssen in den Kustomize-Konfigurationsdateien unter `config/rbac/` definiert werden, nicht direkt in der CSV-Datei. Das `operator-sdk` generiert beim Ausführen von `make bundle` automatisch die CSV-Datei aus diesen Konfigurationen.

## Durchgeführte Änderungen

### 1. Manager ClusterRole aktualisiert
**Datei:** `config/rbac/role.yaml`

Hinzugefügte Berechtigungen für IPRuleConfig:
- `ipruleconfigs/finalizers` - update
- `ipruleconfigs/status` - get, patch, update

Diese werden zu den bereits vorhandenen Berechtigungen für Agents und IPRules hinzugefügt.

### 2. Agent Controller ServiceAccount-Name korrigiert
**Datei:** `internal/controller/agent_controller.go`

Geändert von: `ip-rule-operator-iprule-agent`
Geändert zu: `iprule-agent`

Dies entspricht nun dem ServiceAccount-Namen, der in `config/rbac/agent_service_account.yaml` definiert ist.

## Bestehende RBAC-Struktur

### Controller-Manager Berechtigungen
Die ClusterRole `manager-role` hat folgende Berechtigungen:

```yaml
- Core API (""): 
  - secrets, services: get, list, watch
  - serviceaccounts: create, get, list, watch
  
- api.operator.brtrm.dev:
  - agents: get, list, watch
  - agents/status: get, patch, update
  - agents/finalizers: update
  - iprules: create, delete, get, list, patch, update, watch
  - iprules/status: get, patch, update
  - iprules/finalizers: update
  - ipruleconfigs: create, delete, get, list, patch, update, watch
  - ipruleconfigs/status: get, patch, update
  - ipruleconfigs/finalizers: update
  
- apps:
  - daemonsets: create, delete, get, list, patch, update, watch
  - daemonsets/status: get
  - daemonsets/finalizers: create, delete, get, update
```

### Agent Berechtigungen
Die ClusterRole `iprule-agent-role` hat:

```yaml
- Core API (""):
  - nodes: get, list, watch
  
- api.operator.brtrm.dev:
  - ipruleconfigs, iprules: all verbs (*)
  
- apps/v1:
  - daemonsets/finalizers: update
```

## Bundle neu generieren

Nach den Änderungen in den Config-Dateien muss das Bundle neu generiert werden:

```bash
# 1. Manifests und Bundle generieren
make manifests
make bundle

# 2. Bundle-Image bauen (falls podman/docker verfügbar)
make bundle-build

# 3. Bundle-Image pushen
make bundle-push

# 4. Catalog neu bauen und pushen
make catalog-build catalog-push
```

## Deployment-Prozess

### Für Development (ohne OLM):
```bash
# CRDs installieren
make install

# Controller deployen
make deploy
```

### Für Production (mit OLM):
```bash
# Bundle und Catalog müssen aktualisiert sein
# Dann über OLM installieren oder upgraden
```

## Wichtige Dateien

### RBAC-Konfiguration:
- `config/rbac/role.yaml` - Manager ClusterRole
- `config/rbac/role_binding.yaml` - Manager ClusterRoleBinding
- `config/rbac/service_account.yaml` - Manager ServiceAccount
- `config/rbac/agent_cluster_role.yaml` - Agent ClusterRole
- `config/rbac/agent_cluster_role_binding.yaml` - Agent ClusterRoleBinding
- `config/rbac/agent_service_account.yaml` - Agent ServiceAccount

### Kustomize:
- `config/rbac/kustomization.yaml` - Liste aller RBAC-Ressourcen
- `config/default/kustomization.yaml` - Haupt-Kustomization
- `config/manifests/kustomization.yaml` - OLM Bundle-Kustomization

### Controller:
- `internal/controller/agent_controller.go` - Agent Controller (verwendet ServiceAccount)
- `internal/controller/iprule_controller.go` - IPRule Controller

## Kubebuilder RBAC-Marker

Die RBAC-Berechtigungen werden auch durch Marker in den Controller-Dateien definiert:

```go
// +kubebuilder:rbac:groups=api.operator.brtrm.dev,resources=agents,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
```

Diese Marker werden von `make manifests` verwendet, um `config/rbac/role.yaml` zu generieren.

## Zusammenfassung der Fixes

1. ✅ **RBAC-Berechtigungen in config/rbac/role.yaml hinzugefügt**
   - IPRuleConfig status und finalizers
   
2. ✅ **ServiceAccount-Name im Controller korrigiert**
   - Von `ip-rule-operator-iprule-agent` zu `iprule-agent`

3. ✅ **Korrekte Vorgehensweise dokumentiert**
   - Änderungen in config/ statt direkt in bundle/

## Nächste Schritte

1. **Code committen:**
   ```bash
   git add config/rbac/role.yaml
   git add internal/controller/agent_controller.go
   git commit -m "fix: Add missing RBAC permissions and correct ServiceAccount name"
   ```

2. **Bundle neu generieren:**
   ```bash
   make manifests
   make bundle
   ```

3. **Images bauen und pushen:**
   ```bash
   make docker-build docker-push
   make docker-build-agent docker-push-agent
   make bundle-build bundle-push
   make catalog-build catalog-push
   ```

4. **Operator aktualisieren** (je nach Deployment-Methode)

## Hinweise

- Die CSV-Datei (`bundle/manifests/ip-rule-operator.clusterserviceversion.yaml`) sollte **nicht** manuell bearbeitet werden
- Alle RBAC-Änderungen müssen in `config/rbac/` vorgenommen werden
- Nach jeder Änderung in `config/` muss `make bundle` ausgeführt werden
- Der ServiceAccount-Name muss zwischen Kustomize-Config und Controller-Code übereinstimmen

