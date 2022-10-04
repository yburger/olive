SHELL := /bin/bash

# ==============================================================================
# Building containers

VERSION := 0.5.0

all: olive

olive:
	docker build \
		-f zarf/docker/dockerfile.olive-api \
		-t olive-api-arm64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Modules support

tidy:
	go mod tidy

run:
	go run -ldflags="-X main.build=${VERSION}" app/services/olive-api/main.go

# ==============================================================================
# Running from within k8s/kind

KIND_CLUSTER := olive-cluster

# Upgrade to latest Kind: brew upgrade kind
# For full Kind v0.16 release notes: https://github.com/kubernetes-sigs/kind/releases/tag/v0.16.0
# The image used below was copied by the above link and supports both amd64 and arm64.

kind-up:
	kind create cluster \
		--image kindest/node:v1.25.2@sha256:9be91e9e9cdf116809841fc77ebdb8845443c4c72fe5218f3ae9eb57fdb4bace \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=olive-system

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)
	