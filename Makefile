# -*- Makefile -*-

NAME := facette
VERSION := 0.5.0dev

BUILD_DATE := $(shell date +%F)
BUILD_HASH := $(shell git rev-parse --short HEAD)

PREFIX ?= /usr/local

ENV ?= production

GO ?= vgo
GOLINT ?= golint

GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)

YARN ?= yarn
YARN_ARGS ?= --cwd ui

PANDOC ?= pandoc
PANDOC_ARGS = --standalone --to man

ifeq ($(shell uname -s),Darwin)
TAR ?= gtar
else
TAR ?= tar
endif

GIT_HOOKS := $(patsubst misc/git-hooks/%,.git/hooks/%,$(wildcard misc/git-hooks/*))

BIN_LIST := $(patsubst cmd/%,%,$(wildcard cmd/*))
PKG_LIST = $(call uniq,$(dir $(wildcard */*.go)))
MAN_LIST = $(patsubst docs/man/%.md,%,$(wildcard docs/man/*.[0-9].md))

DIST_DIR ?= dist

tput = $(shell tty 1>/dev/null 2>&1 && tput $1)
print_error = (echo "$(call tput,setaf 1)Error:$(call tput,sgr0) $1")
print_step = echo "$(call tput,setaf 4)***$(call tput,sgr0) $1"
uniq = $(if $1,$(firstword $1) $(call uniq,$(filter-out $(firstword $1),$1)))

all: build

clean:
	@$(call print_step,"Cleaning files...")
	@rm -rf bin/ dist/

build: build-bin build-assets build-docs

ifneq ($(filter builtin_assets,$(TAGS)),)
build-bin: build-assets
	@$(call print_step,"Embedding assets files...")
	@go-bindata \
		-prefix $(DIST_DIR)/assets \
		-tags 'builtin_assets' \
		-o cmd/facette/bindata.go $(DIST_DIR)/assets/...
else
build-bin:
endif
	@$(call print_step,"Building binaries for $(GOOS)/$(GOARCH)...")
	@for bin in $(BIN_LIST); do \
		$(GO) build -i \
			-tags "$(TAGS)" \
			-ldflags "-s -w \
				-X main.version=$(VERSION) \
				-X main.buildDate=$(BUILD_DATE) \
				-X main.buildHash=$(BUILD_HASH) \
			" \
		-o bin/$$bin -v ./cmd/$$bin || ($(call print_error,"failed to build $$bin") && exit 1); \
	done

build-assets: ui/node_modules
	@$(call print_step,"Building assets...")
	@$(YARN) $(YARN_ARGS) build --env $(ENV)

build-docs:
ifneq ($(filter build_docs,$(TAGS)),)
	@$(call print_step,"Generating manual pages...")
	@for man in $(MAN_LIST); do \
		install -d -m 0755 $(DIST_DIR)/man && $(PANDOC) $(PANDOC_ARGS) docs/man/$$man.md >$(DIST_DIR)/man/$$man; \
	done
endif

test: test-bin

test-bin:
	@$(call print_step,"Testing packages...")
	@for pkg in $(PKG_LIST); do \
		$(GO) test -cover -v ./$$pkg || ($(call print_error,"failed to test $$pkg") && exit 1); \
	done

install: install-bin install-assets install-docs

install-bin: build-bin
	@$(call print_step,"Installing binaries...")
	@install -d -m 0755 $(PREFIX)/bin && install -m 0755 bin/* $(PREFIX)/bin/

install-assets: build-assets
	@$(call print_step,"Installing assets...")
	@install -d -m 0755 $(PREFIX)/share/facette && cp -r $(DIST_DIR)/assets $(PREFIX)/share/facette/

install-docs: build-docs
ifneq ($(filter build_docs,$(TAGS)),)
	@$(call print_step,"Installing manual pages...")
	@install -d -m 0755 $(PREFIX)/share/man/man1 && cp -r $(DIST_DIR)/man/* $(PREFIX)/share/man/man1
endif

lint: lint-bin lint-assets

lint-bin:
	@$(call print_step,"Linting binaries and packages...")
	@$(GOLINT) $(BIN_LIST:%=./cmd/%) $(PKG_LIST:%=./%)

lint-assets:
	@$(call print_step,"Checking assets sources...")
	@$(YARN) $(YARN_ARGS) lint

dist: dist-source dist-bin dist-docker

dist-source:
	@$(call print_step,"Building source archive...")
	@install -d -m 0755 $(DIST_DIR) && $(TAR) -czf $(DIST_DIR)/$(NAME)_$(VERSION).tar.gz \
		--transform "flags=r;s/^/$(NAME)-$(VERSION)/" \
		--exclude=.git --exclude=.vscode --exclude=bin --exclude=dist .

dist-bin: build-bin
	@$(call print_step,"Building binary archive...")
	@install -d -m 0755 $(DIST_DIR) && $(TAR) -czf $(DIST_DIR)/$(NAME)_$(VERSION)_$(GOOS)_$(GOARCH).tar.gz \
		--transform "flags=r;s/.*\//$(NAME)-$(VERSION)\//" ./bin/* ./CHANGES.md ./README.md

dist-docker:
	@$(call print_step,"Building Docker image...")
	@docker build -f Dockerfile -t facette/facette:$(VERSION) .

update-locales:
	@$(call print_step,Updating locale files...)
	@$(YARN) $(YARN_ARGS) update-locales

ui/node_modules:
	@$(call print_step,"Fetching node modules...")
	@$(YARN) $(YARN_ARGS)

# Always install missing Git hooks
git-hooks: $(GIT_HOOKS)

.git/hooks/%:
	@$(call print_step,"Installing $* Git hook...")
	@(install -d -m 0755 .git/hooks && cd .git/hooks && ln -s ../../misc/git-hooks/$(@F) .)

-include git-hooks
