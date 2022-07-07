GIT_HASH ?= $(shell git rev-parse HEAD)
GIT_TAG ?= $(shell git describe --tags --always --dirty)

KO_DOCKER_REPO ?= ghcr.io/relengfam/peribolos

.PHONY: ko-build
ko-build:
	ko build --tags $(GIT_TAG),latest --bare \
		--platform=linux/amd64 --image-refs imagerefs \
		github.com/relengfam/peribolos

.PHONY: ko-local
ko-local:
	ko build --local --base-import-paths github.com/relengfam/peribolos

.PHONY: build-sign-images
build-sign-images: ko-build
	GIT_HASH=$(GIT_HASH) GIT_TAG=$(GIT_TAG) ./scripts/sign-images.sh

# kubernetes/org make targets
# TODO(k-org): Remove once integrated into peribolos

SHELL := /usr/bin/env bash

# available for override
GITHUB_TOKEN_PATH ?=

# intentionally hardcoded list to ensure it's high friction to remove someone
ADMINS = cblecker fejta idvoretskyi mrbobbytables nikhita spiffxp
ORGS = $(shell find ./config -type d -mindepth 1 -maxdepth 1 | cut -d/ -f3)

# use absolute path to ./_output, which is .gitignored
OUTPUT_DIR := $(shell pwd)/_output
OUTPUT_BIN_DIR := $(OUTPUT_DIR)/bin

PERIBOLOS_CMD := $(OUTPUT_BIN_DIR)/peribolos
MERGE_CMD := $(PERIBOLOS_CMD) merge

CONFIG_FILES = $(shell find config/ -type f -name '*.yaml')
MERGED_CONFIG := $(OUTPUT_DIR)/gen-config.yaml

# convenience targets for humans
.PHONY: clean
clean:
	rm -rf $(OUTPUT_DIR)

.PHONY: merge
merge: $(MERGE_CMD)

.PHONY: config
config: $(MERGED_CONFIG)

.PHONY: peribolos
peribolos: $(PERIBOLOS_CMD)

# TODO(merge): Fix this test
.PHONY: test
test: config
	go test ./... --config=$(MERGED_CONFIG)

.PHONY: verify
verify:
	./hack/verify.sh

.PHONY: update-prep
update-prep: config peribolos # TODO(merge): Fix this test "config test peribolos"

.PHONY: deploy # --confirm
deploy:
	./admin-update.sh
		$(-*-command-variables-*-) $(filter-out $@,$(MAKECMDGOALS))

# actual targets that only get built if they don't already exist
$(MERGED_CONFIG): $(MERGE_CMD) $(CONFIG_FILES)
	mkdir -p "$(OUTPUT_DIR)"
	$(MERGE_CMD) \
		--merge-teams \
		$(shell for o in $(ORGS); do echo "--org-part=$$o=config/$$o/org.yaml"; done) \
		> $(MERGED_CONFIG)

$(PERIBOLOS_CMD):
	go build -v -o $(PERIBOLOS_CMD)
