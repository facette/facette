# -*- Makefile -*-

BUILD_DIR = build

TEMP_DIR = tmp

PREFIX ?= $(BUILD_DIR)/facette

mesg_start = echo "$(shell tput setaf 4)$(1):$(shell tput sgr0) $(2)"
mesg_step = echo "$(1)"
mesg_ok = echo "result: $(shell tput setaf 2)ok$(shell tput sgr0)"
mesg_fail = (echo "result: $(shell tput setaf 1)fail$(shell tput sgr0)" && false)

# Go
GOPATH = $(realpath $(TEMP_DIR))
export GOPATH

GO ?= go

# Utilities
GOLINT ?= golint
GOLINT_ARGS =

PANDOC ?= pandoc
PANDOC_ARGS = --standalone --to man

UGLIFYJS ?= uglifyjs
UGLIFYSCRIPT_ARGS = --comments --compress --mangle --screw-ie8

JSHINT ?= jshint
JSHINT_ARGS = --show-non-errors

LESSC ?= lessc
LESSC_ARGS = --no-color

all: install

clean:
	@$(call mesg_start,main,Cleaning temporary files...)
	@rm -rf $(TEMP_DIR) && \
		$(call mesg_ok) || $(call mesg_fail)

.PHONY: build
build: build-bin build-man build-static

install: install-bin install-man install-static

lint: lint-bin lint-static

test: test-pkg test-server

$(TEMP_DIR)/src/github.com/facette/facette:
	@$(call mesg_start,main,Creating source symlink...)
	@mkdir -p $(TEMP_DIR)/src/github.com/facette && \
		ln -s ../../../.. $(TEMP_DIR)/src/github.com/facette/facette && \
		$(call mesg_ok) || $(call mesg_fail)

# Binaries
BIN_SRC = $(wildcard cmd/*/*.go)

BIN_OUTPUT = $(addprefix $(TEMP_DIR)/bin/, $(notdir $(wildcard cmd/*)))

PKG_SRC = $(wildcard pkg/*/*.go)

$(BIN_OUTPUT): $(PKG_SRC) $(BIN_SRC) $(TEMP_DIR)/src/github.com/facette/facette
	@$(call mesg_start,$(notdir $@),Building $(notdir $@)...)
	@install -d -m 0755 $(dir $@) && $(GO) build -o $@ cmd/$(notdir $@)/*.go && \
		$(call mesg_ok) || $(call mesg_fail)
	@test ! -f cmd/$(notdir $@)/Makefile || make --no-print-directory -C cmd/$(notdir $@) build

build-bin: $(BIN_OUTPUT)

install-bin: build-bin
	@$(call mesg_start,install,Installing binaries...)
	@install -d -m 0755 $(PREFIX) && cp -Rp $(TEMP_DIR)/bin $(PREFIX) && \
		$(call mesg_ok) || $(call mesg_fail)

lint-bin: build-bin
	@$(call mesg_start,lint,Checking sources with Golint...)
	@$(GOLINT) $(GOLINT_ARGS) cmd pkg && $(call mesg_ok) || $(call mesg_fail)

# Manuals
MAN_SRC = $(wildcard docs/man/*.[0-9].md)

MAN_OUTPUT = $(addprefix $(TEMP_DIR)/man/, $(notdir $(MAN_SRC:.md=)))

$(MAN_OUTPUT): $(MAN_SRC)
	@$(call mesg_start,docs,Generating $(notdir $@) manual page...)
	@install -d -m 0755 $(dir $@) && $(PANDOC) $(PANDOC_ARGS) docs/man/$(notdir $@).md >$@ && \
		$(call mesg_ok) || $(call mesg_fail)

build-man: $(MAN_OUTPUT)

install-man: build-man
	@$(call mesg_start,install,Installing manuals files...)
	@install -d -m 0755 $(PREFIX)/share && cp -Rp $(TEMP_DIR)/man $(PREFIX)/share && \
		$(call mesg_ok) || $(call mesg_fail)

# Static
SCRIPT_SRC = cmd/facette/js/intro.js \
	cmd/facette/js/define.js \
	cmd/facette/js/extend.js \
	cmd/facette/js/utils.js \
	cmd/facette/js/setup.js \
	cmd/facette/js/overlay.js \
	cmd/facette/js/list.js \
	cmd/facette/js/tree.js \
	cmd/facette/js/tooltip.js \
	cmd/facette/js/menu.js \
	cmd/facette/js/input.js \
	cmd/facette/js/select.js \
	cmd/facette/js/link.js \
	cmd/facette/js/admin/intro.js \
	cmd/facette/js/admin/admin.js \
	cmd/facette/js/admin/item.js \
	cmd/facette/js/admin/graph.js \
	cmd/facette/js/admin/collection.js \
	cmd/facette/js/admin/group.js \
	cmd/facette/js/admin/catalog.js \
	cmd/facette/js/admin/outro.js \
	cmd/facette/js/browse/intro.js \
	cmd/facette/js/browse/browse.js \
	cmd/facette/js/browse/collection.js \
	cmd/facette/js/browse/outro.js \
	cmd/facette/js/item.js \
	cmd/facette/js/graph.js \
	cmd/facette/js/stats.js \
	cmd/facette/js/pane.js \
	cmd/facette/js/i18n.js \
	cmd/facette/js/outro.js

SCRIPT_OUTPUT = $(TEMP_DIR)/static/facette.js

MESG_SRC = cmd/facette/js/messages.json

MESG_OUTPUT = $(TEMP_DIR)/static/$(MESG_SRC)

STYLE_SRC = cmd/facette/style/intro.less \
	cmd/facette/style/define.less \
	cmd/facette/style/font.less \
	cmd/facette/style/common.less \
	cmd/facette/style/icon.less \
	cmd/facette/style/form.less \
	cmd/facette/style/tooltip.less \
	cmd/facette/style/date.less \
	cmd/facette/style/overlay.less \
	cmd/facette/style/list.less \
	cmd/facette/style/stats.less \
	cmd/facette/style/admin.less \
	cmd/facette/style/graph.less

STYLE_OUTPUT = $(TEMP_DIR)/static/style.css

STYLE_PRINT_SRC = cmd/facette/style/intro.less \
	cmd/facette/style/define.less \
	cmd/facette/style/common-print.less

STYLE_PRINT_OUTPUT = $(TEMP_DIR)/static/style.print.css

HTML_SRC = cmd/facette/html/common \
	cmd/facette/html/admin \
	cmd/facette/html/browse \
	cmd/facette/html/error.html \
	cmd/facette/html/wait.html \
	cmd/facette/html/layout.html

HTML_OUTPUT = $(TEMP_DIR)/static/html

$(SCRIPT_OUTPUT): $(SCRIPT_SRC)
	@$(call mesg_start,static,Merging script files into $(notdir $(SCRIPT_OUTPUT:.js=.src.js))...)
	@install -d -m 0755 $(TEMP_DIR)/static && cat $(SCRIPT_SRC) >$(SCRIPT_OUTPUT:.js=.src.js) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,static,Packing $(notdir $(SCRIPT_OUTPUT:.js=.src.js)) file...)
	@$(UGLIFYJS) $(UGLIFYSCRIPT_ARGS) --output $(SCRIPT_OUTPUT) $(SCRIPT_OUTPUT:.js=.src.js) && \
		$(call mesg_ok) || $(call mesg_fail)

$(MESG_OUTPUT): $(MESG_SRC)
	@$(call mesg_start,static,Packing $(MESG_SRC) file...)
	@install -d -m 0755 $(TEMP_DIR)/static && \
		sed -e 's/^\s\+//g;s/\s\+$$//g' $(MESG_SRC) | sed -e ':a;N;s/\n//;ta' >$(MESG_OUTPUT) && \
		$(call mesg_ok) || $(call mesg_fail)

$(STYLE_OUTPUT): $(STYLE_SRC)
	@$(call mesg_start,static,Merging style files into $(notdir $(STYLE_OUTPUT:.css=.src.css))...)
	@install -d -m 0755 $(TEMP_DIR)/static && cat $(STYLE_SRC) >$(STYLE_OUTPUT:.css=.src.css) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,static,Packing $(notdir $(STYLE_OUTPUT:.css=.src.css)) file...)
	@$(LESSC) $(LESSC_ARGS) --yui-compress $(STYLE_OUTPUT:.css=.src.css) >$(STYLE_OUTPUT) && \
		$(call mesg_ok) || $(call mesg_fail)

$(STYLE_PRINT_OUTPUT): $(STYLE_PRINT_SRC)
	@$(call mesg_start,static,Merging style files into $(notdir $(STYLE_PRINT_OUTPUT:.css=.src.css))...)
	@install -d -m 0755 $(TEMP_DIR)/static && cat $(STYLE_PRINT_SRC) >$(STYLE_PRINT_OUTPUT:.css=.src.css) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,static,Packing $(notdir $(STYLE_PRINT_OUTPUT:.css=.src.css)) file...)
	@$(LESSC) $(LESSC_ARGS) --yui-compress $(STYLE_PRINT_OUTPUT:.css=.src.css) >$(STYLE_PRINT_OUTPUT) && \
		$(call mesg_ok) || $(call mesg_fail)

$(HTML_OUTPUT): $(HTML_SRC)
	@$(call mesg_start,static,Copying HTML files...)
	@install -d -m 0755 $(HTML_OUTPUT) && cp -r $(HTML_SRC) $(HTML_OUTPUT)/ && \
		$(call mesg_ok) || $(call mesg_fail)

build-static: $(SCRIPT_OUTPUT) $(MESG_FILE) $(STYLE_OUTPUT) $(STYLE_PRINT_OUTPUT) $(HTML_OUTPUT)

install-static: build-static
	@$(call mesg_start,install,Installing static files...)
	@install -d -m 0755 $(PREFIX)/share && cp -Rp $(SCRIPT_OUTPUT) $(STYLE_OUTPUT) $(HTML_OUTPUT) $(PREFIX)/share && \
		$(call mesg_ok) || $(call mesg_fail)

lint-static: $(SCRIPT_OUTPUT)
	@$(call mesg_start,lint,Checking $(notdir $(SCRIPT_OUTPUT:.js=.src.js)) with JSHint...)
	-@$(JSHINT) $(JSHINT_ARGS) $(SCRIPT_OUTPUT:.js=.src.js) && \
		$(call mesg_ok) || $(call mesg_fail)

# Test
PKG_SRC = $(wildcard pkg/*)

test-pkg:
	@install -d -m 0755 $(TEMP_DIR)/tests && (cd $(TEMP_DIR)/tests; for ENTRY in $(PKG_SRC); do \
		$(call mesg_start,test,Testing $$ENTRY package...); \
		$(GO) test -c -i ../../$$ENTRY && \
			(test ! -f ./`basename $$ENTRY`.test || ./`basename $$ENTRY`.test -test.v=true) && \
			$(call mesg_ok) || $(call mesg_fail); \
	done)

test-server: build-bin
	@$(call mesg_start,test,Starting facette server...)
	@install -d -m 0755 $(TEMP_DIR)/tests && ($(TEMP_DIR)/bin/facette -c tests/facette.json >/dev/null &) && \
		$(call mesg_ok) || $(call mesg_fail)

	@$(call mesg_start,test,Running server tests...)
	@(cd $(TEMP_DIR)/tests; $(GO) test -c -i ../../cmd/facette) && \
		./$(TEMP_DIR)/tests/facette.test -test.v=true -c tests/facette.json && \
		$(call mesg_ok) || (kill -2 `cat $(TEMP_DIR)/tests/facette.pid`; $(call mesg_fail))

	@$(call mesg_start,test,Stopping facette server...)
	@kill -2 `cat $(TEMP_DIR)/tests/facette.pid` && \
		$(call mesg_ok) || $(call mesg_fail)
