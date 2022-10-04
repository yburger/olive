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