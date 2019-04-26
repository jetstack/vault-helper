# Copyright Jetstack Ltd. See LICENSE for details.
PACKAGE_NAME ?= github.com/jetstack/vault-helper
CONTAINER_DIR := /go/src/$(PACKAGE_NAME)
GO_VERSION := 1.11.4

BINDIR ?= $(CURDIR)/bin
PATH   := $(BINDIR):$(PATH)

HACK_DIR     ?= hack

REGISTRY := quay.io/jetstack
IMAGE_NAME := vault-helper
IMAGE_TAGS := canary
BUILD_TAG := build

CI_COMMIT_TAG ?= $(shell git rev-parse HEAD)
CI_COMMIT_SHA ?= unknown

VAULT_VERSION := 0.9.6
VAULT_HASH := 3f1f346ff7aaf367fed6a3e83e5a07fdc032f22860585e36c3674f9ead08dbaf

help:
	# all       - runs verify, build targets
	# test      - runs go_test target
	# build     - runs generate, and then go_build targets
	# generate  - generates mocks and assets files
	# verify    - verifies generated files & scripts
	# image     - build docker image

.PHONY: all test verify

verify: generate go_verify

all: verify build

build: generate go_build

test: go_generate go_test

generate: go_generate

go_verify: go_fmt go_vet verify_boilerplate go_test

go_test:
	go test --count=1 $$(go list ./pkg/... ./cmd/...)

go_fmt:
	@set -e; \
	GO_FMT=$$(git ls-files *.go | grep -v 'vendor/' | xargs gofmt -d); \
	if [ -n "$${GO_FMT}" ] ; then \
		echo "Please run go fmt"; \
		echo "$$GO_FMT"; \
		exit 1; \
	fi

go_vet:
	go vet $$(go list ./pkg/... ./cmd/...)

verify_boilerplate:
	$(HACK_DIR)/verify-boilerplate.sh

go_build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%d_%H:%M:%S)' -o vault-helper_linux_amd64

image:
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(BUILD_TAG) .

image_push: image
	set -e; \
		for tag in $(IMAGE_TAGS); do \
		docker tag $(REGISTRY)/$(IMAGE_NAME):$(BUILD_TAG) $(REGISTRY)/$(IMAGE_NAME):$${tag} ; \
		docker push $(REGISTRY)/$(IMAGE_NAME):$${tag}; \
		done

bin/mockgen:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/mockgen ./vendor/github.com/golang/mock/mockgen

bin/vault:
	which unzip || ( apt-get update && apt-get -y install unzip)
	mkdir -p $(BINDIR)
	curl -sL  https://releases.hashicorp.com/vault/$(VAULT_VERSION)/vault_$(VAULT_VERSION)_linux_amd64.zip > $(BINDIR)/vault.zip
	echo "$(VAULT_HASH)  $(BINDIR)/vault.zip" | sha256sum  -c
	cd $(BINDIR) && unzip vault.zip
	rm $(BINDIR)/vault.zip

depend: bin/mockgen bin/vault

go_generate: depend
	$(BINDIR)/mockgen -package kubernetes -source=pkg/kubernetes/kubernetes.go > pkg/kubernetes/kubernetes_mocks_test.go
	sed -i '1s/^/\/\/ Copyright Jetstack Ltd. See LICENSE for details.\n/' pkg/kubernetes/kubernetes_mocks_test.go

# Builder image targets
#######################
docker_%:
	# create a container
	$(eval CONTAINER_ID := $(shell docker create \
		-i \
		-w $(CONTAINER_DIR) \
		golang:${GO_VERSION} \
		/bin/bash -c "make $*" \
	))

	# copy stuff into container
	(git ls-files && git ls-files --others --exclude-standard) | tar cf -  -T - | docker cp - $(CONTAINER_ID):$(CONTAINER_DIR)

	# run build inside container
	docker start -a -i $(CONTAINER_ID)

	# copy artifacts over
	if [ "$*" = "build" ]; then \
		docker cp $(CONTAINER_ID):$(CONTAINER_DIR)/vault-helper_linux_amd64 vault-helper_linux_amd64; \
	fi

	# remove container
	docker rm $(CONTAINER_ID)

