# HttpRepeater local build helpers (releases are built on GitHub Actions; see .github/workflows/release.yml).
#
# Version badge at build time:
#   git describe --tags --dirty   -> v0.2.0 or v0.2.0-3-gabc1234-dirty
#   no tag                        -> dev+<short-commit>
#   no git                        -> empty (badge hidden)

SHELL := /bin/sh

APP := HttpRepeater
BIN_DIR := $(CURDIR)/build/bin
OUT := $(CURDIR)/build/release

VERSION ?= $(shell git describe --tags --dirty 2>/dev/null || true)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || true)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

ifeq ($(strip $(VERSION)),)
DISPLAY_VERSION := $(if $(strip $(COMMIT)),dev+$(COMMIT),)
else
DISPLAY_VERSION := $(VERSION)
endif

PLATFORM ?=
WAILS_BUILD := wails build $(if $(strip $(PLATFORM)),-platform $(PLATFORM),)

.PHONY: help dev build test clean version

help:
	@echo "HttpRepeater (local):"
	@echo "  make dev [PLATFORM=...]     wails dev with a git-derived version"
	@echo "  make build [PLATFORM=...]   wails build (output in build/bin/)"
	@echo "  make test                   go test ./... + frontend build"
	@echo "  make version                print computed version metadata"
	@echo "  make clean                  remove build/bin and build/release"
	@echo ""
	@echo "Releases (macOS + Windows zip): push a tag, e.g. git tag v1.2.0 && git push origin v1.2.0"
	@echo "  Workflow: .github/workflows/release.yml"

version:
	@echo "version: $(if $(strip $(DISPLAY_VERSION)),$(DISPLAY_VERSION),unknown)"
	@echo "commit:  $(if $(strip $(COMMIT)),$(COMMIT),unknown)"
	@echo "built:   $(BUILD_DATE)"

dev:
	VITE_APP_VERSION="$(DISPLAY_VERSION)" \
	VITE_APP_COMMIT="$(COMMIT)" \
	VITE_APP_BUILD_DATE="$(BUILD_DATE)" \
	wails dev $(if $(strip $(PLATFORM)),-platform $(PLATFORM),)

build:
	@echo "Building $(APP) $(if $(strip $(DISPLAY_VERSION)),$(DISPLAY_VERSION),without version)..."
	VITE_APP_VERSION="$(DISPLAY_VERSION)" \
	VITE_APP_COMMIT="$(COMMIT)" \
	VITE_APP_BUILD_DATE="$(BUILD_DATE)" \
	$(WAILS_BUILD)

test:
	go test ./...
	cd frontend && npm run build

clean:
	rm -rf "$(OUT)" "$(BIN_DIR)"
