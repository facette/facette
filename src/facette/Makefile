#!/usr/bin/make -f
# -*- Makefile -*-

include ../../Makefile.inc

GIT = git
HG = hg

GO = go
GOPATH = $(realpath $(BUILD_DIR)):$(realpath ../..)
export GOPATH

GOLINT = golint
GOLINT_ARGS =

BIN_FILES = facette \
	facettectl

SRC_FILES = $(wildcard */*.go)

PKG_TESTS = facette/server \
	facette/utils

all: lint install

clean:
	@$(call mesg_start,Cleaning built files...)
	@(test ! -d $(PREFIX)/bin || (cd $(PREFIX)/bin && rm -rf $(BIN_FILES))) && $(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,Cleaning fetched dependencies...)
	@rm -rf $(BUILD_DIR)/src && $(call mesg_ok) || $(call mesg_fail)
	@$(call clean_build)

lint: $(SRC_FILES)
	@$(call mesg_start,Checking sources with Golint...)
	-@$(GOLINT) $(GOLINT_ARGS) . && $(call mesg_ok) || $(call mesg_fail)

depends:
	@$(call mesg_start,Fetching Go dependencies...)
	@$(MAKEFILE_DIR)/fetch-depends.sh $(BUILD_DIR)/src && $(call mesg_ok) || $(call mesg_fail)

$(BIN_FILES): depends $(SRC_FILES)
	@$(call mesg_start,Building $@ binary file...)
	@install -m 0755 -d $(PREFIX)/bin && \
		$(GO) build -o $(PREFIX)/bin/$@ cmd/$@/*.go && \
		$(call mesg_ok) || $(call mesg_fail)

build: $(BIN_FILES)

test: build
	@$(call mesg_start,Starting facette server...)
	@install -m 0755 -d $(BUILD_DIR)/tests && (cd $(BUILD_DIR)/tests && \
		$(PREFIX)/bin/facette -c ../../tests/facette.json >/dev/null & echo "$$!" >$(BUILD_DIR)/facette.pid) && \
		$(call mesg_ok) || $(call mesg_fail)

	@cd $(BUILD_DIR)/tests && for ENTRY in $(PKG_TESTS); do \
		$(call mesg_start,Running $$ENTRY tests...); \
		$(GO) test -c -i $$ENTRY && ./`basename $$ENTRY`.test -test.v -c ../../tests/facette.json && \
			$(call mesg_ok) || $(call mesg_fail); \
		done

	@$(call mesg_start,Stopping facette server...)
	@kill `cat $(BUILD_DIR)/facette.pid` && rm -f $(BUILD_DIR)/facette.pid && $(call mesg_ok) || $(call mesg_fail)

install: build

devel: install
