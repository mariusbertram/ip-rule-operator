# Merged Makefile (inkl. FBC Development Targets)
# ORIGINAL INHALT + hinzugefügter Abschnitt "FBC Development" am Ende.
# Wenn zufrieden: vorhandenes Makefile durch diese Datei ersetzen (umbenennen zu Makefile).

# VERSION defines the project version for the bundle.
VERSION ?= 0.0.1

# CHANNELS / DEFAULT_CHANNEL
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

IMAGE_TAG_BASE ?= ghcr.io/mariusbertram/ip-rule-operator
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)
BUNDLE_GEN_FLAGS ?= -q --overwrite --extra-service-accounts iprule-agent --version $(VERSION) $(BUNDLE_METADATA_OPTS)
USE_IMAGE_DIGESTS ?= true
ifeq ($(USE_IMAGE_DIGESTS), true)
  BUNDLE_GEN_FLAGS += --use-image-digests
endif

OPERATOR_SDK_VERSION ?= v1.41.1
IMG ?= ghcr.io/mariusbertram/iprule-controller:v$(VERSION)
AGENT_IMG ?= ghcr.io/mariusbertram/iprule-agent:v$(VERSION)
AGENT_IMG_LATEST ?= ghcr.io/mariusbertram/iprule-agent:latest
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

CONTAINER_TOOL ?= podman
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS=":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0,5) }' $(MAKEFILE_LIST)

##@ Development
.PHONY: manifests
manifests: controller-gen ## Generate CRDs and RBAC.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate deepcopy code.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## go fmt
	go fmt ./...

.PHONY: vet
vet: ## go vet
	go vet ./...

.PHONY: test
test: manifests generate fmt vet setup-envtest ## Run unit tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test $$(go list ./... | grep -v /e2e) -coverprofile cover.out

KIND_CLUSTER ?= ip-rule-operator-test-e2e
.PHONY: setup-test-e2e
setup-test-e2e: ## Create kind cluster if missing
	@command -v $(KIND) >/dev/null 2>&1 || { echo "Kind not installed"; exit 1; }
	@case "$$($(KIND) get clusters)" in *"$(KIND_CLUSTER)"*) echo "Kind cluster exists" ;; *) echo "Creating kind cluster"; $(KIND) create cluster --name $(KIND_CLUSTER) ;; esac

.PHONY: test-e2e
test-e2e: setup-test-e2e manifests generate fmt vet ## Run e2e tests
	KIND_CLUSTER=$(KIND_CLUSTER) go test ./test/e2e/ -v -ginkgo.v
	$(MAKE) cleanup-test-e2e

.PHONY: cleanup-test-e2e
cleanup-test-e2e: ## Delete kind cluster
	- $(KIND) delete cluster --name $(KIND_CLUSTER)

.PHONY: lint
lint: golangci-lint ## Run linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Lint with fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Validate lint config
	$(GOLANGCI_LINT) config verify

##@ Build
.PHONY: agent-build
agent-build: generate fmt vet ## Build agent
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/iprule-agent ./cmd/agent

.PHONY: agent-image-build
agent-image-build: agent-build ## Agent image
	$(CONTAINER_TOOL) build -f Dockerfile.agent -t $(AGENT_IMG) -t $(AGENT_IMG_LATEST) .

.PHONY: agent-image-push
agent-image-push: ## Push agent image
	$(CONTAINER_TOOL) push $(AGENT_IMG) && $(CONTAINER_TOOL) push $(AGENT_IMG_LATEST)

.PHONY: build
build: manifests generate fmt vet ## Build controller
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run locally
	go run ./cmd/main.go

PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-build
docker-build: ## Build manager image
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push manager image
	$(CONTAINER_TOOL) push ${IMG}

.PHONY: docker-buildx
docker-buildx: ## Multi-arch build
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name ip-rule-operator-builder
	$(CONTAINER_TOOL) buildx use ip-rule-operator-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm ip-rule-operator-builder
	rm Dockerfile.cross

.PHONY: build-installer
build-installer: manifests generate kustomize ## Aggregate install yaml
	mkdir -p dist
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > dist/install.yaml

##@ Deployment
ifndef ignore-not-found
  ignore-not-found = false
endif
.PHONY: install
install: manifests kustomize ## Install CRDs
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -
	mv config/manager/manager.yaml.orig config/manager/manager.yaml || true
	rm -f config/manager/manager.yaml.bak || true

.PHONY: undeploy
undeploy: kustomize ## Remove controller
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Dependencies
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

KUBECTL ?= kubectl
KIND ?= kind
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

KUSTOMIZE_VERSION ?= v5.6.0
CONTROLLER_TOOLS_VERSION ?= v0.18.0
ENVTEST_VERSION ?= $(shell go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime | awk -F'[v.]' '{printf "release-%d.%d", $$2, $$3}')
ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')
GOLANGCI_LINT_VERSION ?= v2.1.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: setup-envtest
setup-envtest: envtest
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@$(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path || { echo "envtest setup failed"; exit 1; }

.PHONY: envtest
envtest: $(ENVTEST)
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

.PHONY: operator-sdk
OPERATOR_SDK ?= $(LOCALBIN)/operator-sdk
operator-sdk:
ifeq (,$(wildcard $(OPERATOR_SDK)))
ifeq (, $(shell which operator-sdk 2>/dev/null))
	@{ set -e; mkdir -p $(dir $(OPERATOR_SDK)); OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && curl -sSLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH}; chmod +x $(OPERATOR_SDK); }
else
OPERATOR_SDK = $(shell which operator-sdk)
endif
endif

.PHONY: bundle
bundle: manifests kustomize operator-sdk
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle $(BUNDLE_GEN_FLAGS)
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: bundle-build
bundle-build:
	$(CONTAINER_TOOL) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push:
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = $(LOCALBIN)/opm
opm:
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ set -e; mkdir -p $(dir $(OPM)); OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.55.0/$${OS}-$${ARCH}-opm; chmod +x $(OPM); }
else
OPM = $(shell which opm)
endif
endif

# Catalog index targets
BUNDLE_IMGS ?= $(BUNDLE_IMG)
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif
.PHONY: catalog-build
catalog-build: opm
	$(OPM) index add --container-tool $(CONTAINER_TOOL) --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

.PHONY: catalog-push
catalog-push:
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

.PHONY: build-all
build-all: docker-build agent-image-build

.PHONY: push-all
push-all: docker-push agent-image-push bundle bundle-build bundle-push

# Package manifests (legacy style)
ifneq ($(origin FROM_VERSION), undefined)
PKG_FROM_VERSION := --from-version=$(FROM_VERSION)
endif
ifneq ($(origin CHANNEL), undefined)
PKG_CHANNELS := --channel=$(CHANNEL)
endif
ifeq ($(IS_CHANNEL_DEFAULT), 1)
PKG_IS_DEFAULT_CHANNEL := --default-channel
endif
PKG_MAN_OPTS ?= $(PKG_FROM_VERSION) $(PKG_CHANNELS) $(PKG_IS_DEFAULT_CHANNEL)
.PHONY: packagemanifests
packagemanifests: kustomize manifests
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate packagemanifests -q --version $(VERSION) $(PKG_MAN_OPTS)

##@ FBC Development (merged)
FBC_IMG ?= $(IMAGE_TAG_BASE)-fbc:v$(VERSION)
FBC_DEV_DIR ?= catalog/dev
FBC_DEV_FILE ?= $(FBC_DEV_DIR)/ip-rule-operator-v$(VERSION)-fbc.yaml
FBC_PACKAGE ?= ip-rule-operator
FBC_CHANNEL ?= stable-v0

.PHONY: fbc-dev
fbc-dev: bundle opm ## Generate dev File-Based Catalog YAML (not for production)
	@echo "[FBC] Generating FBC YAML ./bundle -> $(FBC_DEV_FILE) (version $(VERSION))"
	@mkdir -p $(FBC_DEV_DIR)
	$(OPM) render ./bundle --output=yaml > $(FBC_DEV_FILE).tmp
	@{ \
	 echo "---"; echo "name: $(FBC_PACKAGE)"; echo "schema: olm.package"; echo "defaultChannel: $(FBC_CHANNEL)"; \
	 echo "---"; echo "schema: olm.channel"; echo "name: $(FBC_CHANNEL)"; echo "package: $(FBC_PACKAGE)"; echo "entries:"; \
	 echo "- name: $(FBC_PACKAGE).v$(VERSION)"; } > $(FBC_DEV_FILE)
	cat $(FBC_DEV_FILE).tmp >> $(FBC_DEV_FILE)
	@rm -f $(FBC_DEV_FILE).tmp
	@echo "[FBC] Wrote $(FBC_DEV_FILE)"

.PHONY: fbc-dev-validate
fbc-dev-validate: fbc-dev opm ## Validate dev FBC directory
	$(OPM) validate $(FBC_DEV_DIR)
	@echo "[FBC] Validation OK"

.PHONY: fbc-dev-image
fbc-dev-image: fbc-dev-validate ## Build dev FBC image
	@echo "[FBC] Generating Dockerfile for dev FBC"
	$(OPM) generate dockerfile $(FBC_DEV_DIR)
	$(CONTAINER_TOOL) build -f $(FBC_DEV_DIR)/Dockerfile -t $(FBC_IMG) $(FBC_DEV_DIR)
	@echo "[FBC] Built image $(FBC_IMG)"

.PHONY: fbc-dev-push
fbc-dev-push: fbc-dev-image ## Push dev FBC image
	$(CONTAINER_TOOL) push $(FBC_IMG)

.PHONY: fbc-dev-clean
fbc-dev-clean: ## Remove dev FBC artifacts
	rm -f $(FBC_DEV_DIR)/Dockerfile || true
	rm -f $(FBC_DEV_FILE) || true
	@echo "[FBC] Cleaned dev FBC artifacts"

