# HttpRepeater build and release helpers.
#
# The version is derived from git at build time:
#   git describe --tags --dirty   -> v0.2.0 or v0.2.0-3-gabc1234-dirty
#   no tag                        -> dev+<short-commit>
#   no git                        -> empty (the version badge is hidden)
#
# Release flow:
#   make notes VERSION=v0.2.0      # generates build/release/notes.md
#   make release VERSION=v0.2.0    # build + package + tag + GitHub release
#
# `make release` requires a clean working tree and an authenticated `gh` CLI.

SHELL := /bin/sh

APP := HttpRepeater
BIN := httpRepeater
BIN_DIR := $(CURDIR)/build/bin
OUT := $(CURDIR)/build/release

VERSION ?= $(shell git describe --tags --dirty 2>/dev/null || true)
VERSION_NUM := $(patsubst v%,%,$(VERSION))
ifeq ($(strip $(VERSION_NUM)),0.0.0)
override VERSION :=
override VERSION_NUM :=
endif
TAG := v$(VERSION_NUM)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || true)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

ifeq ($(strip $(VERSION)),)
DISPLAY_VERSION := $(if $(strip $(COMMIT)),dev+$(COMMIT),)
else
DISPLAY_VERSION := $(VERSION)
endif

PKG_VERSION := $(if $(strip $(VERSION_NUM)),$(VERSION_NUM),dev)
PLATFORM ?=
WAILS_BUILD := wails build $(if $(strip $(PLATFORM)),-platform $(PLATFORM),)

ifeq ($(strip $(PLATFORM)),)
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
TARGET_OS := darwin
else ifeq ($(UNAME_S),Linux)
TARGET_OS := linux
else ifneq (,$(findstring MINGW,$(UNAME_S)))
TARGET_OS := windows
else ifneq (,$(findstring MSYS,$(UNAME_S)))
TARGET_OS := windows
else ifneq (,$(findstring CYGWIN,$(UNAME_S)))
TARGET_OS := windows
else
TARGET_OS := unknown
endif
TARGET_ARCH := $(shell uname -m)
else ifeq (,$(findstring /,$(PLATFORM)))
$(error PLATFORM must be os/arch, e.g. PLATFORM=windows/amd64)
else
TARGET_OS := $(firstword $(subst /, ,$(PLATFORM)))
TARGET_ARCH := $(lastword $(subst /, ,$(PLATFORM)))
endif

.PHONY: help dev build test package package-assets notes release publish build-release check-version check-release clean version

help:
	@echo "HttpRepeater:"
	@echo "  make dev [PLATFORM=...]             wails dev with a git-derived version"
	@echo "  make build [PLATFORM=...]           wails build with a git-derived version"
	@echo "  make test                           go test ./... + frontend build"
	@echo "  make package [PLATFORM=...]         build and package for the current target"
	@echo "  make notes VERSION=v0.2.0           generate build/release/notes.md"
	@echo "  make release VERSION=v0.2.0         build, package, tag and publish a release"
	@echo "  make publish VERSION=v0.2.0         tag and publish existing artifacts"
	@echo "  make version                        print the computed version"
	@echo "  make clean                          remove build/release and build/bin"

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

package:
	@$(MAKE) --no-print-directory build VERSION="$(VERSION)" PLATFORM="$(PLATFORM)"
	@$(MAKE) --no-print-directory package-assets VERSION="$(VERSION)" PLATFORM="$(PLATFORM)"

package-assets:
	@mkdir -p "$(OUT)"
	@os="$(TARGET_OS)"; arch="$(TARGET_ARCH)"; label="$$os"; \
	if [ "$$os" = darwin ]; then label=macos; fi; \
	case "$$os" in \
		darwin) \
			src="$(BIN_DIR)/$(BIN).app"; \
			[ -d "$$src" ] || { echo "missing $$src; build for $$os first" >&2; exit 1; }; \
			out="$(OUT)/$(APP)-$(PKG_VERSION)-$$label-$$arch.zip"; \
			ditto -c -k --norsrc --keepParent "$$src" "$$out" ;; \
		windows) \
			src="$(BIN_DIR)/$(BIN).exe"; \
			[ -f "$$src" ] || { echo "missing $$src; build for $$os first" >&2; exit 1; }; \
			out="$(OUT)/$(APP)-$(PKG_VERSION)-$$label-$$arch.zip"; \
			if command -v zip >/dev/null 2>&1; then \
				(cd "$(BIN_DIR)" && zip -q -r "$$out" "$(BIN).exe"); \
			else \
				powershell -NoProfile -Command "Compress-Archive -Path '$(BIN_DIR)/$(BIN).exe' -DestinationPath '$$out' -Force"; \
			fi ;; \
		linux) \
			src="$(BIN_DIR)/$(BIN)"; \
			[ -f "$$src" ] || { echo "missing $$src; build for $$os first" >&2; exit 1; }; \
			out="$(OUT)/$(APP)-$(PKG_VERSION)-$$label-$$arch.tar.gz"; \
			tar -C "$(BIN_DIR)" -czf "$$out" "$(BIN)" ;; \
		*) echo "unsupported platform; set PLATFORM=os/arch" >&2; exit 1 ;; \
	esac; \
	echo "Packaged $$out"

notes:
	@$(MAKE) --no-print-directory check-version VERSION="$(VERSION)"
	@bash scripts/release-notes.sh "$(TAG)" "$(OUT)/notes.md"

release:
	@$(MAKE) --no-print-directory build-release VERSION="$(VERSION)" PLATFORM="$(PLATFORM)"
	@$(MAKE) --no-print-directory package-assets VERSION="$(VERSION)" PLATFORM="$(PLATFORM)"
	@test -f "$(OUT)/notes.md" || bash scripts/release-notes.sh "$(TAG)" "$(OUT)/notes.md"
	@$(MAKE) --no-print-directory publish VERSION="$(VERSION)" PLATFORM="$(PLATFORM)"

publish:
	@$(MAKE) --no-print-directory check-release VERSION="$(VERSION)"
	@test -f "$(OUT)/notes.md" || { echo "missing $(OUT)/notes.md; run 'make notes VERSION=$(TAG)' first" >&2; exit 1; }
	@assets="$$(find "$(OUT)" -maxdepth 1 -type f \( -name "$(APP)-$(PKG_VERSION)-*.zip" -o -name "$(APP)-$(PKG_VERSION)-*.tar.gz" \) | sort)"; \
	[ -n "$$assets" ] || { echo "no artifacts for $(PKG_VERSION); run 'make package-assets' first" >&2; exit 1; }; \
	if [ "$(DRY_RUN)" = "1" ]; then \
		echo "Dry run: tag $(TAG), push, and publish these assets:"; echo "$$assets"; \
		exit 0; \
	fi; \
	command -v gh >/dev/null 2>&1 || { echo "gh CLI is required: https://cli.github.com/" >&2; exit 1; }; \
	gh auth status >/dev/null 2>&1 || { echo "gh is not authenticated; run 'gh auth login'" >&2; exit 1; }; \
	if git rev-parse "$(TAG)" >/dev/null 2>&1; then \
		test "$$(git rev-list -n 1 "$(TAG)")" = "$$(git rev-parse HEAD)" || { echo "tag $(TAG) exists on another commit" >&2; exit 1; }; \
		if ! git ls-remote --tags origin "refs/tags/$(TAG)" | grep -q .; then \
			git push origin "$(TAG)"; \
		fi; \
	else \
		git tag -a "$(TAG)" -m "$(APP) $(TAG)"; \
		git push origin "$(TAG)"; \
	fi; \
	if gh release view "$(TAG)" >/dev/null 2>&1; then \
		gh release upload "$(TAG)" $$assets --clobber; \
		gh release edit "$(TAG)" --title "$(APP) $(TAG)" --notes-file "$(OUT)/notes.md"; \
	else \
		gh release create "$(TAG)" $$assets --title "$(APP) $(TAG)" --notes-file "$(OUT)/notes.md"; \
	fi

build-release:
	@$(MAKE) --no-print-directory check-release VERSION="$(VERSION)"
	@node scripts/wails-version.mjs set "$(VERSION_NUM)"
	@trap 'node scripts/wails-version.mjs restore' EXIT INT TERM; \
		VITE_APP_VERSION="$(VERSION)" \
		VITE_APP_COMMIT="$(COMMIT)" \
		VITE_APP_BUILD_DATE="$(BUILD_DATE)" \
		$(WAILS_BUILD)

check-version:
	@test -n "$(strip $(VERSION))" || { echo "VERSION is required, e.g. VERSION=v0.2.0" >&2; exit 1; }
	@echo "$(VERSION_NUM)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$$' || { echo "VERSION must be semver, e.g. v0.2.0" >&2; exit 1; }
	@git rev-parse --git-dir >/dev/null 2>&1 || { echo "not a git repository" >&2; exit 1; }

check-release:
	@$(MAKE) --no-print-directory check-version VERSION="$(VERSION)"
	@git remote get-url origin >/dev/null 2>&1 || { echo "no 'origin' remote configured" >&2; exit 1; }
	@test -z "$$(git status --porcelain)" || { echo "working tree is dirty; commit or stash first" >&2; exit 1; }

clean:
	rm -rf "$(OUT)" "$(BIN_DIR)"
