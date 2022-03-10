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

imagerefs := $(shell cat imagerefs)
sign-refs := $(foreach ref,$(imagerefs),$(ref))
.PHONY: sign-images
sign-images:
	cosign sign -a GIT_TAG=$(GIT_TAG) -a GIT_HASH=$(GIT_HASH) $(sign-refs)
