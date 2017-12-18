# -*- Makefile -*-

VERSION := 0.4.0

BUILD_DATE := $(shell date +%F)
BUILD_HASH := $(shell git rev-parse --short HEAD)

PREFIX ?= /usr/local

GO ?= go

GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)

BUILD_NAME = facette-$(GOOS)-$(GOARCH)
BUILD_DIR = build/$(BUILD_NAME)
BUILD_ENV ?= production

export GOPATH = $(realpath $(BUILD_DIR)):$(realpath $(BUILD_DIR))/vendor

GOLINT ?= golint
GOLINT_ARGS =

NPM ?= npm
NPM_ARGS =

GULP ?= node_modules/.bin/gulp
GULP_ARGS = --no-color

PANDOC ?= pandoc
PANDOC_ARGS = --standalone --to man

BIN_LIST = $(patsubst src/cmd/%,%,$(wildcard src/cmd/*))
PKG_LIST = $(patsubst src/%,%,$(wildcard src/facette/*))
MAN_LIST = $(patsubst docs/man/%.md,%,$(wildcard docs/man/*.[0-9].md))

mesg_start = echo "$(shell tty -s && tput setaf 4)$(1):$(shell tty -s && tput sgr0) $(2)"
mesg_step = echo "$(1)"
mesg_ok = echo "result: $(shell tty -s && tput setaf 2)ok$(shell tty -s && tput sgr0)"
mesg_fail = (echo "result: $(shell tty -s && tput setaf 1)fail$(shell tty -s && tput sgr0)" && false)

all: build

clean:
	@$(call mesg_start,clean,Removing build data...)
	@rm -rf $(BUILD_DIR) src/cmd/facette/bindata.go && \
		$(call mesg_ok) || $(call mesg_fail)
	@rmdir build 2>/dev/null || true

clean-all: clean
	@$(call mesg_start,clean,Removing assets build dependencies...)
	@rm -rf node_modules && \
		$(call mesg_ok) || $(call mesg_fail)

build: build-bin build-assets build-docs

build-dir:
	@$(call mesg_start,build,Preparing build directory...)
	@(install -d -m 0755 $(BUILD_DIR)/bin $(BUILD_DIR)/pkg $(BUILD_DIR)/vendor && \
		(cd $(BUILD_DIR) && ln -sf ../../src .) && (cd $(BUILD_DIR)/vendor && ln -sf ../../../vendor/src ../pkg .)) && \
		$(call mesg_ok) || $(call mesg_fail)

ifneq ($(filter builtin_assets,$(BUILD_TAGS)),)
build-bin: build-dir build-assets
	@$(call mesg_start,build,Embedding assets files...)
	@go-bindata \
			-prefix $(BUILD_DIR)/assets \
			-tags 'builtin_assets' \
			-o src/cmd/facette/bindata.go $(BUILD_DIR)/assets/... && \
		$(call mesg_ok) || $(call mesg_fail)
else
build-bin: build-dir
endif
	@$(call mesg_start,build,Building binaries...)
	@(for bin in $(BIN_LIST); do \
		$(GO) build -i -v \
			-tags "$(BUILD_TAGS)" \
			-ldflags "-s -w \
				-X main.version=$(VERSION) \
				-X main.buildDate=$(BUILD_DATE) \
				-X main.buildHash=$(BUILD_HASH) \
			" \
			-o $(BUILD_DIR)/bin/$$bin ./src/cmd/$$bin || exit 1; \
	done) && $(call mesg_ok) || $(call mesg_fail)

build-assets: node_modules
	@$(call mesg_start,build,Building assets...)
	@BUILD_DIR=$(BUILD_DIR) $(GULP) $(GULP_ARGS) build --env $(BUILD_ENV) >/dev/null && \
		$(call mesg_ok) || $(call mesg_fail)

build-docs:
ifneq ($(filter build_docs,$(BUILD_TAGS)),)
	@for man in $(MAN_LIST); do \
		$(call mesg_start,docs,Generating $$man manual page...); \
		install -d -m 0755 $(BUILD_DIR)/man && $(PANDOC) $(PANDOC_ARGS) docs/man/$$man.md >$(BUILD_DIR)/man/$$man && \
			$(call mesg_ok) || $(call mesg_fail); \
	done
endif

test: test-bin

test-bin: build-dir
	@$(call mesg_start,test,Testing packages...)
	@(cd $(BUILD_DIR) && for pkg in $(PKG_LIST); do \
		install -d -m 0755 tests/`dirname $$pkg`; \
		$(GO) test -v \
			-tags "$(BUILD_TAGS)" \
			-coverprofile tests/$$pkg.out $$pkg \
			|| exit 1; \
		test ! -f tests/$$pkg.out || $(GO) tool cover -o tests/$$pkg.func -func=tests/$$pkg.out || exit 1; \
	done) && $(call mesg_ok) || $(call mesg_fail)

install: install-bin install-assets install-docs

install-bin: build-bin
	@$(call mesg_start,install,Installing binaries...)
	@install -d -m 0755 $(PREFIX)/bin && install -m 0755 $(BUILD_DIR)/bin/* $(PREFIX)/bin/ && \
		$(call mesg_ok) || $(call mesg_fail)

install-assets: build-assets
	@$(call mesg_start,install,Installing assets...)
	@install -d -m 0755 $(PREFIX)/share/facette && cp -r $(BUILD_DIR)/assets $(PREFIX)/share/facette/ && \
		$(call mesg_ok) || $(call mesg_fail)

install-docs: build-docs
ifneq ($(filter build_docs,$(BUILD_TAGS)),)
	@$(call mesg_start,install,Installing manual pages...)
	@install -d -m 0755 $(PREFIX)/share/man/man1 && cp -r $(BUILD_DIR)/man/* $(PREFIX)/share/man/man1 && \
		$(call mesg_ok) || $(call mesg_fail)
endif

lint: lint-bin lint-assets

lint-bin:
	@for pkg in $(PKG_LIST) $(BIN_LIST:%=cmd/%); do \
		$(call mesg_start,lint,Checking $$pkg sources...); \
		$(GOLINT) $(GOLINT_ARGS) ./src/$$pkg && \
			$(call mesg_ok) || $(call mesg_fail); \
	done

lint-assets:
	@$(call mesg_start,lint,Checking assets sources...)
	@BUILD_DIR=$(BUILD_DIR) $(GULP) $(GULP_ARGS) lint && \
		$(call mesg_ok) || $(call mesg_fail)

update-locales:
	@$(call mesg_start,locale,Updating locale files...)
	@BUILD_DIR=$(BUILD_DIR) $(GULP) $(GULP_ARGS) update-locales >/dev/null && \
		$(call mesg_ok) || $(call mesg_fail)

node_modules:
	@$(call mesg_start,build,Retrieving assets build dependencies...)
	@$(NPM) $(NPM_ARGS) install --package-lock=false >/dev/null && \
		$(call mesg_ok) || $(call mesg_fail)
