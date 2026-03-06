MODULE   := github.com/usegradient/gradient
BINARY   := gradient
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X main.Version=$(VERSION)
BUILD_DIR := dist

# ── Build ────────────────────────────────────────────────────────

.PHONY: build
build: ## Build for the current platform
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) .

.PHONY: build-all
build-all: ## Cross-compile for all release platforms
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		out=$(BUILD_DIR)/$(BINARY)_$${os}_$${arch}; \
		echo "  building $$out"; \
		GOOS=$$os GOARCH=$$arch go build -ldflags '$(LDFLAGS)' -o $$out . || exit 1; \
	done
	@echo "✓ binaries in $(BUILD_DIR)/"

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR) $(BINARY)

# ── Release ──────────────────────────────────────────────────────
#
#   make release-patch   v0.2.3 → v0.2.4
#   make release-minor   v0.2.3 → v0.3.0
#   make release-major   v0.2.3 → v1.0.0
#
# Each target: bumps version, tags, cross-compiles, pushes tag,
# creates a GitHub release with all binaries attached.

LATEST_TAG := $(shell git tag --list 'v*' --sort=-v:refname | head -1)

# Parse semver components from the latest tag (default v0.0.0)
_MAJOR = $(shell echo "$(or $(LATEST_TAG),v0.0.0)" | sed 's/^v//' | cut -d. -f1)
_MINOR = $(shell echo "$(or $(LATEST_TAG),v0.0.0)" | sed 's/^v//' | cut -d. -f2)
_PATCH = $(shell echo "$(or $(LATEST_TAG),v0.0.0)" | sed 's/^v//' | cut -d. -f3)

NEXT_PATCH := v$(_MAJOR).$(_MINOR).$(shell echo $$(( $(_PATCH) + 1 )))
NEXT_MINOR := v$(_MAJOR).$(shell echo $$(( $(_MINOR) + 1 ))).0
NEXT_MAJOR := v$(shell echo $$(( $(_MAJOR) + 1 ))).0.0

define do_release
	@echo "── releasing $(1) (previous: $(LATEST_TAG)) ──"
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "✗ Working tree is dirty. Commit or stash changes first."; exit 1; \
	fi
	git tag -a $(1) -m "Release $(1)"
	$(MAKE) build-all VERSION=$(1)
	git push origin $(1)
	gh release create $(1) $(BUILD_DIR)/* \
		--title "$(1)" \
		--generate-notes
	@echo "✓ $(1) released — https://github.com/use-gradient/gradient/releases/tag/$(1)"
endef

.PHONY: release-patch
release-patch: ## Bump patch (x.y.Z+1), tag, build, and publish release
	$(call do_release,$(NEXT_PATCH))

.PHONY: release-minor
release-minor: ## Bump minor (x.Y+1.0), tag, build, and publish release
	$(call do_release,$(NEXT_MINOR))

.PHONY: release-major
release-major: ## Bump major (X+1.0.0), tag, build, and publish release
	$(call do_release,$(NEXT_MAJOR))

# ── Helpers ──────────────────────────────────────────────────────

.PHONY: version
version: ## Print current and next versions
	@echo "latest tag : $(or $(LATEST_TAG),(none))"
	@echo "next patch : $(NEXT_PATCH)"
	@echo "next minor : $(NEXT_MINOR)"
	@echo "next major : $(NEXT_MAJOR)"

.PHONY: help
help: ## Show this help
	@grep -E '^[a-z_-]+:.*## ' $(MAKEFILE_LIST) | \
		awk -F ':.*## ' '{printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
