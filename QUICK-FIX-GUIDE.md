# Schnellstart: RBAC-Probleme beheben

## Problem
Der Operator meldet: `failed to wait for agent caches to sync`

## Sofort-Fix (für laufenden Operator)

Wenn der Operator bereits deployed ist und RBAC-Fehler auftreten:

```bash
# 1. Fix-RBAC anwenden (temporär)
kubectl apply -f fix-rbac.yaml

# 2. Controller neu starten
kubectl delete pod -n ip-rule -l control-plane=controller-manager
```

## Dauerhafte Lösung (für neue Deployments)

### 1. Änderungen wurden bereits gemacht in:
- ✅ `config/rbac/role.yaml` - IPRuleConfig Berechtigungen hinzugefügt
- ✅ `internal/controller/agent_controller.go` - ServiceAccount-Name korrigiert

### 2. Bundle neu generieren:

**Option A: Mit WSL (empfohlen):**
```bash
wsl -d RedHatEnterpriseLinux-10.0 -e zsh -c "cd /mnt/c/Users/mariu/GolandProjects/ip-rule-operator && make bundle"
```

**Option B: Direkt in Linux/WSL:**
```bash
cd /mnt/c/Users/mariu/GolandProjects/ip-rule-operator
make manifests
make bundle
```

### 3. Änderungen committen:
```bash
git add config/rbac/role.yaml
git add internal/controller/agent_controller.go
git add bundle/
git commit -m "fix: Add missing RBAC permissions for IPRuleConfig and correct agent ServiceAccount name"
```

### 4. Images neu bauen und pushen:
```bash
# Controller Image
make docker-build docker-push

# Agent Image
make docker-build-agent docker-push-agent

# Bundle und Catalog
make bundle-build bundle-push
make catalog-build catalog-push
```

### 5. Operator neu deployen:
```bash
# Via OLM (Production)
# Der Catalog wird automatisch aktualisiert

# Oder direkt (Development)
make deploy
```

## Verifikation

```bash
# 1. Prüfe, ob Pods laufen
kubectl get pods -n ip-rule

# 2. Prüfe Logs auf Fehler
kubectl logs -n ip-rule -l control-plane=controller-manager --tail=50

# 3. Prüfe Agent DaemonSet
kubectl get daemonset -n ip-rule iprule-agent

# 4. Prüfe Agent-Status
kubectl get agent -n ip-rule agent-sample -o yaml
```

## Erwartetes Ergebnis

✅ Keine RBAC-Fehler in den Logs
✅ Controller-Manager läuft stabil
✅ Agent DaemonSet wird erstellt
✅ Agent-Pods werden gestartet (wenn Nodes verfügbar)

## Bei Problemen

Siehe ausführliche Dokumentation in:
- `RBAC-KONFIGURATION.md` - Technische Details
- `RBAC-FIX-DOCUMENTATION.md` - Problem-Analyse

