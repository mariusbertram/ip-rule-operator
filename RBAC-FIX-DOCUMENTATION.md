# Lösung: IP-Rule-Operator RBAC-Fehler

## Problem
Der Controller-Manager im Namespace `ip-rule` konnte nicht starten und meldete den Fehler:
```
failed to wait for agent caches to sync kind source: *v1alpha1.Agent: timed out waiting for cache to be synced for Kind *v1alpha1.Agent
```

## Ursache
Die vom OLM (Operator Lifecycle Manager) generierte ClusterRole hatte nicht die notwendigen Berechtigungen für:
- Agents (api.operator.brtrm.dev)
- IPRules (api.operator.brtrm.dev)
- IPRuleConfigs (api.operator.brtrm.dev)
- DaemonSets (apps)
- Services (core)
- ServiceAccounts (core)
- Secrets (core)

## Lösung

### 1. RBAC-Berechtigungen hinzugefügt
Eine neue ClusterRole `ip-rule-manager-role` wurde erstellt mit allen notwendigen Berechtigungen:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ip-rule-manager-role
rules:
  - apiGroups: [""]
    resources: [secrets, serviceaccounts, services]
    verbs: [get, list, watch, create]
  
  - apiGroups: [api.operator.brtrm.dev]
    resources: [agents, iprules, ipruleconfigs]
    verbs: [get, list, watch]
  
  - apiGroups: [api.operator.brtrm.dev]
    resources: [agents/status, agents/finalizers, iprules/status, iprules/finalizers, ipruleconfigs/status, ipruleconfigs/finalizers]
    verbs: [get, update, patch]
  
  - apiGroups: [apps]
    resources: [daemonsets]
    verbs: [get, list, watch, create, update, patch, delete]
  
  - apiGroups: [apps]
    resources: [daemonsets/status, daemonsets/finalizers]
    verbs: [get, update]
  
  - apiGroups: [authentication.k8s.io]
    resources: [tokenreviews]
    verbs: [create]
  
  - apiGroups: [authorization.k8s.io]
    resources: [subjectaccessreviews]
    verbs: [create]
```

### 2. ClusterRoleBinding erstellt
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ip-rule-manager-rolebinding-fix
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ip-rule-manager-role
subjects:
  - kind: ServiceAccount
    name: ip-rule-controller-manager
    namespace: ip-rule
```

### 3. ServiceAccount für Agent erstellt
```bash
kubectl create serviceaccount ip-rule-operator-iprule-agent -n ip-rule
```

### 4. CSV-Datei aktualisiert
Die Datei `bundle/manifests/ip-rule-operator.clusterserviceversion.yaml` wurde aktualisiert, damit zukünftige Deployments die korrekten Berechtigungen haben.

## Anwendung der Lösung

```bash
# 1. RBAC anwenden
kubectl apply -f fix-rbac.yaml

# 2. Controller-Manager neu starten
kubectl delete pod -n ip-rule -l control-plane=controller-manager

# 3. ServiceAccount erstellen
kubectl create serviceaccount ip-rule-operator-iprule-agent -n ip-rule
```

## Ergebnis
✅ Der Controller-Manager startet erfolgreich  
✅ Agent-Caches synchronisieren sich korrekt  
✅ DaemonSet `iprule-agent` wird erstellt  
✅ Keine RBAC-Fehler mehr in den Logs  

## Dateien geändert
1. `bundle/manifests/ip-rule-operator.clusterserviceversion.yaml` - RBAC-Berechtigungen aktualisiert
2. `fix-rbac.yaml` - Temporäre RBAC-Fix-Datei erstellt

## Nächste Schritte für Production
1. Bundle neu bauen mit `make bundle`
2. Bundle-Image neu bauen und pushen
3. Catalog neu erstellen und deployen
4. Operator über OLM neu installieren

## Notizen
- Der Fehler trat auf, weil OLM nur die explizit in der CSV definierten Berechtigungen erteilt
- Die RBAC-Marker (`+kubebuilder:rbac`) in den Controller-Dateien werden nur beim Bundle-Build verwendet
- Bei manuellen Deployments (ohne OLM) müssen die RBAC-Manifeste aus `config/rbac` separat angewendet werden

