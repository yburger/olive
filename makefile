SHELL := /bin/bash

VERSION := 0.5.0

tidy:
	go mod tidy

run:
	go run -ldflags="-X main.build=${VERSION}" app/services/olive-api/main.go