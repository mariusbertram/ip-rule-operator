# IP Rule Operator - Air-Gapped Images Kurzreferenz

## Benötigte Production Images

```
ghcr.io/mariusbertram/iprule-controller:v0.0.1
ghcr.io/mariusbertram/iprule-agent:v0.0.1
```

## Bundle Generation (mit Digests)

```bash
make bundle \
  IMG=ghcr.io/mariusbertram/iprule-controller:v0.0.1 \
  AGENT_IMG=ghcr.io/mariusbertram/iprule-agent:v0.0.1 \
  VERSION=0.0.1 \
  USE_IMAGE_DIGESTS=true
```

## CSV relatedImages (automatisch generiert)

```yaml
spec:
  relatedImages:
    - image: ghcr.io/mariusbertram/iprule-controller@sha256:3b86a4a96a421aa75e46d0c9e6988dfba4cf1cf304309f65ec1637239cd104f2
      name: manager
    - image: ghcr.io/mariusbertram/iprule-agent@sha256:<DIGEST>
      name: agent
```

## Mirroring für Air-Gapped

```bash
SOURCE="ghcr.io/mariusbertram"
TARGET="registry.airgap.local/ip-rule-operator"
VERSION="v0.0.1"

# Controller
skopeo copy --all \
  docker://${SOURCE}/iprule-controller:${VERSION} \
  docker://${TARGET}/iprule-controller:${VERSION}

# Agent
skopeo copy --all \
  docker://${SOURCE}/iprule-agent:${VERSION} \
  docker://${TARGET}/iprule-agent:${VERSION}
```

## OpenShift ICSP

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

## Digest Resolution Tools

1. **skopeo** (empfohlen): `skopeo inspect docker://IMAGE`
2. **docker**: `docker inspect IMAGE`
3. **podman**: `podman inspect IMAGE`

## Vollständige Dokumentation

Siehe: [docs/AIRGAP-IMAGES.md](./AIRGAP-IMAGES.md)

