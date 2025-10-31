# RBAC und `make manifests` - Lösungen

## Problem
Der Befehl `make manifests` nutzt `controller-gen` um RBAC-Regeln automatisch aus den Controller-Annotationen zu generieren und überschreibt dabei die generierten Dateien. Manuell erstellte RBAC-Dateien in `config/rbac/` werden ignoriert oder überschrieben.

## Empfohlene Lösungen

### Lösung 1: Separate Verzeichnisse (EMPFOHLEN)
Die beste Lösung ist, manuell verwaltete RBAC-Dateien in einem separaten Verzeichnis zu speichern.

#### Struktur:
```
config/
  rbac/                      # Automatisch generiert durch controller-gen
    role.yaml
    role_binding.yaml
    ...
  rbac-custom/              # Manuell verwaltete RBAC
    agent_cluster_role.yaml
    agent_scc_binding.yaml
    ...
  default/
    kustomization.yaml      # Beide Verzeichnisse einbinden
```

#### Vorteile:
- Klare Trennung zwischen generierten und manuellen Ressourcen
- Keine Konflikte mit `make manifests`
- Einfache Wartung

---

### Lösung 2: RBAC Markers in Go Code verwenden
Für Agent-spezifische RBAC können Sie kubebuilder-Markers im Code verwenden.

#### Beispiel in `cmd/agent/main.go`:
```go
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups=api.operator.brtrm.dev,resources=ipruleconfigs,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=api.operator.brtrm.dev,resources=ipruleconfigs/status,verbs=get;update;patch

package main
```

Dann generiert `controller-gen` diese automatisch.

#### Vorteile:
- RBAC-Regeln sind im Code dokumentiert
- Automatisch synchronisiert mit dem Code
- Keine manuellen YAML-Dateien

#### Nachteile:
- Weniger Flexibilität für komplexe RBAC
- SCC (SecurityContextConstraints) müssen trotzdem manuell sein

---

### Lösung 3: Protected Files mit eigener Kustomization
Erstellen Sie eine geschützte Kustomization-Ebene.

#### In `config/rbac-protected/kustomization.yaml`:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- agent_cluster_role.yaml
- agent_scc_binding.yaml
- agent_service_account.yaml
```

#### In `config/default/kustomization.yaml`:
```yaml
resources:
- ../rbac
- ../rbac-protected  # Zusätzlich einbinden
```

---

### Lösung 4: Post-Generation Script
Erstellen Sie ein Script, das nach `make manifests` ausgeführt wird.

#### `hack/restore-rbac.sh`:
```bash
#!/bin/bash
# Restore custom RBAC files after controller-gen

cp config/rbac-backup/agent_* config/rbac/
```

#### Im Makefile:
```makefile
.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	@echo "Restoring custom RBAC files..."
	@bash hack/restore-rbac.sh
```

---

## Meine Empfehlung für Ihr Projekt

Ich empfehle **Lösung 1** mit einer kleinen Anpassung:

### Implementierung:

1. **Verzeichnisstruktur:**
```
config/
  rbac/                          # Nur generierte Dateien
  rbac-agent/                    # Agent-spezifische RBAC
    agent_cluster_role.yaml
    agent_scc_binding.yaml
    agent-scc-clusterrole.yaml
    agent_service_account.yaml
    kustomization.yaml
```

2. **Update `config/default/kustomization.yaml`:**
```yaml
resources:
- ../crd
- ../rbac
- ../rbac-agent           # Neue Zeile
- ../manager
```

3. **Makefile bleibt unverändert:**
```makefile
.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
```

### Warum diese Lösung?

1. ✅ **Keine Konflikte** - `controller-gen` überschreibt nur `config/rbac/`
2. ✅ **Klare Trennung** - Entwickler wissen sofort, wo was hingehört
3. ✅ **Wartbar** - Einfaches Hinzufügen neuer RBAC-Regeln
4. ✅ **Minimale Änderungen** - Nur Dateien verschieben und Kustomization anpassen
5. ✅ **GitOps-freundlich** - Klare Historie welche Dateien manuell/automatisch sind

---

## Alternative für einfachere Projekte

Falls Sie nur wenige Agent-RBAC-Regeln haben, könnten Sie auch **Lösung 2** verwenden und die RBAC-Markers direkt in `cmd/agent/main.go` einfügen. Dies hält alles an einem Ort.

Für OpenShift SCC (SecurityContextConstraints) müssten Sie aber trotzdem manuelle YAML-Dateien haben, da diese nicht von kubebuilder generiert werden können.

---

## Nächste Schritte

Möchten Sie, dass ich:
1. Die Verzeichnisstruktur entsprechend Lösung 1 umorganisiere?
2. RBAC-Markers in den Agent-Code einfüge (Lösung 2)?
3. Eine andere der vorgeschlagenen Lösungen implementiere?

