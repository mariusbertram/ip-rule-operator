# Additional File-Based Catalog (FBC) development targets
# Usage examples:
#   make -f Makefile -f Makefile.fbc-dev.mk fbc-dev
#   make -f Makefile -f Makefile.fbc-dev.mk fbc-dev-validate
#   make -f Makefile -f Makefile.fbc-dev.mk fbc-dev-image
# These targets are for local development only.

FBC_IMG ?= $(IMAGE_TAG_BASE)-fbc:v$(VERSION)
FBC_DEV_DIR ?= catalog/dev
FBC_DEV_FILE ?= $(FBC_DEV_DIR)/ip-rule-operator-v$(VERSION)-fbc.yaml
# Package & Channel Metadata (dev)
FBC_PACKAGE ?= ip-rule-operator
FBC_CHANNEL ?= stable-v0

##@ FBC Development (extra include)
.PHONY: fbc-dev
fbc-dev: bundle opm ## Render a development File-Based Catalog YAML from the local bundle directory (adds package+channel)
	@echo "[FBC] Generating FBC YAML ./bundle -> $(FBC_DEV_FILE) (version $(VERSION))"
	@mkdir -p $(FBC_DEV_DIR)
	# Render bundle objects to temp
	$(OPM) render ./bundle --output=yaml > $(FBC_DEV_FILE).tmp
	# Write package + channel header
	@{ \
	  echo "---"; \
	  echo "name: $(FBC_PACKAGE)"; \
	  echo "schema: olm.package"; \
	  echo "defaultChannel: $(FBC_CHANNEL)"; \
	  echo "---"; \
	  echo "schema: olm.channel"; \
	  echo "name: $(FBC_CHANNEL)"; \
	  echo "package: $(FBC_PACKAGE)"; \
	  echo "entries:"; \
	  echo "- name: $(FBC_PACKAGE).v$(VERSION)"; \
	} > $(FBC_DEV_FILE)
	# Append rendered bundle content
	cat $(FBC_DEV_FILE).tmp >> $(FBC_DEV_FILE)
	@rm -f $(FBC_DEV_FILE).tmp
	@echo "[FBC] Wrote $(FBC_DEV_FILE)"

.PHONY: fbc-dev-validate
fbc-dev-validate: fbc-dev opm ## Validate the generated development FBC YAML
	$(OPM) validate $(FBC_DEV_DIR)
	@echo "[FBC] Validation OK"

.PHONY: fbc-dev-image
fbc-dev-image: fbc-dev-validate ## Build a dev-only FBC image (not for production)
	@echo "[FBC] Generating Dockerfile for dev FBC"
	$(OPM) generate dockerfile $(FBC_DEV_DIR)
	$(CONTAINER_TOOL) build -f $(FBC_DEV_DIR)/Dockerfile -t $(FBC_IMG) $(FBC_DEV_DIR)
	@echo "[FBC] Built image $(FBC_IMG)"

.PHONY: fbc-dev-push
fbc-dev-push: fbc-dev-image ## Push the dev FBC image (for internal testing only)
	$(CONTAINER_TOOL) push $(FBC_IMG)

.PHONY: fbc-dev-clean
fbc-dev-clean: ## Remove generated dev FBC artifacts
	rm -f $(FBC_DEV_DIR)/Dockerfile || true
	rm -f $(FBC_DEV_FILE) || true
	@echo "[FBC] Cleaned development FBC artifacts"
