# Version variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# LDFLAGS for version injection
LDFLAGS = -ldflags="-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)' -w -s"

# Platform variables
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

help:
	@echo "Please use 'make <target>' where <target> is one of the following:"
	@echo "  test                 to run unit tests."
	@echo "  build                to build the app as a binary."
	@echo "  build-all            to build the app for multiple platforms."
	@echo "  build-image          to build the app container."
	@echo "  run                  to run the app with go."
	@echo "  version              to show version information."

run:
	go run $(LDFLAGS) ./cmd/calais/main.go

test:
	go test -v -coverpkg=./... -coverprofile=profile.cov ./...
	go tool cover -func profile.cov
	rm profile.cov

build:
	@mkdir -p ./bin
	CGO_ENABLED=0 go build -mod=readonly $(LDFLAGS) -o ./bin/calais ./cmd/calais/*

build-all:
	@mkdir -p ./bin
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		echo "Building for $$GOOS/$$GOARCH..."; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build -mod=readonly $(LDFLAGS) -o ./bin/calais-$$GOOS-$$GOARCH ./cmd/calais/*; \
		done

build-image:
	docker build -t atmosx/calais:latest .
	@if [ "$(VERSION)" != "dev" ]; then docker tag atmosx/calais:latest atmosx/calais:$(VERSION); fi

version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"


.PHONY: help run test build build-all build-image version
