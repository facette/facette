# -*- Makefile -*-

NAME := facette
VERSION := 0.5.0dev
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
REVISION := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date +"%F %T")

REPO_PATH := facette.io/facette

PREFIX ?= /usr/local

ENV ?= production

ifeq ($(ENV),production)
override TAGS += builtin_assets
endif

GO ?= go
GOLINT ?= golint

GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)

YARN ?= yarn
YARN_ARGS ?= --emoji false --no-color --cwd ui

PANDOC ?= pandoc
PANDOC_ARGS := --standalone --to man

ifeq ($(shell uname -s),Darwin)
TAR ?= gtar
else
TAR ?= tar
endif

GIT_HOOKS := $(patsubst misc/git-hooks/%,.git/hooks/%,$(wildcard misc/git-hooks/*))

BIN_LIST := $(patsubst cmd/%,%,$(wildcard cmd/*))
PKG_LIST := $(call uniq,$(dir $(wildcard */*.go)))
MAN_LIST := $(patsubst docs/man/%.md,%,$(wildcard docs/man/*.[0-9].md))
UI_LIST := $(shell find ui/src -type f)

DIST_DIR ?= dist

tput = $(shell tty 1>/dev/null 2>&1 && tput $1)
print_step = echo "$(call tput,setaf 4)***$(call tput,sgr0) $1"
uniq = $(if $1,$(firstword $1) $(call uniq,$(filter-out $(firstword $1),$1)))

all: build

clean:
	@$(call print_step,"Cleaning files...")
	@rm -rf bin/ dist/ web/bindata.go

build: build-bin build-assets build-docs

ifneq ($(filter builtin_assets,$(TAGS)),)
build-bin: build-assets
else
build-bin:
endif
	@$(call print_step,"Building binaries for $(GOOS)/$(GOARCH)...")
	@$(GO) generate -tags "$(TAGS)" ./... && for bin in $(BIN_LIST); do \
		$(GO) build -i \
			-tags "$(TAGS)" \
			-ldflags "-s -w \
				-X '$(REPO_PATH)/version.Version=$(VERSION)' \
				-X '$(REPO_PATH)/version.Branch=$(BRANCH)' \
				-X '$(REPO_PATH)/version.Revision=$(REVISION)' \
				-X '$(REPO_PATH)/version.BuildDate=$(BUILD_DATE)' \
			" \
		-o bin/$$bin -v ./cmd/$$bin || exit 1; \
	done

build-assets: ui/node_modules
	@$(call print_step,"Building assets...")
	@rm -rf $(DIST_DIR)/assets/ && $(YARN) $(YARN_ARGS) build --env $(ENV)

build-docs:
ifeq ($(filter skip_docs,$(TAGS)),)
	@$(call print_step,"Generating manual pages...")
	@for man in $(MAN_LIST); do \
		echo $$man; \
		install -d -m 0755 $(DIST_DIR)/man && $(PANDOC) $(PANDOC_ARGS) docs/man/$$man.md >$(DIST_DIR)/man/$$man; \
	done
endif

test: test-bin

test-bin:
	@$(call print_step,"Testing packages...")
	@for pkg in $(PKG_LIST); do \
		$(GO) test -cover -tags "$(TAGS)" -v ./$$pkg || exit 1; \
	done

install: install-bin install-docs

install-bin: build-bin
	@$(call print_step,"Installing binaries...")
	@install -d -m 0755 $(PREFIX)/bin && install -m 0755 bin/* $(PREFIX)/bin/

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
		--transform "flags=r;s|^\./|$(NAME)-$(VERSION)/|" \
		--exclude=.git --exclude=.vscode --exclude=bin --exclude=bindata.go --exclude=dist \
		--exclude=node_modules --exclude=var .

dist-bin: build-bin
	@$(call print_step,"Building binary archive...")
	@install -d -m 0755 $(DIST_DIR) && $(TAR) -czf $(DIST_DIR)/$(NAME)_$(VERSION)_$(GOOS)_$(GOARCH).tar.gz \
		--transform "flags=r;s/.*\//$(NAME)-$(VERSION)\//" ./bin/* ./CHANGES.md ./README.md

dist-deb:
	@$(call print_step,"Building Debian packages...")
	@./misc/scripts/build-debian.sh

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
