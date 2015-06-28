# -*- Makefile -*-

VERSION := 0.4.0dev

BUILD_DATE := $(shell date +%F)

TAGS ?= facette \
	graphite \
	kairosdb \
	influxdb \
	rrd

PREFIX ?= /usr/local

UNAME := $(shell uname -s)

GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)

BUILD_NAME = facette-$(GOOS)-$(GOARCH)
BUILD_DIR = build/$(BUILD_NAME)

GOPATH = $(realpath $(BUILD_DIR))
export GOPATH

GO ?= go

mesg_start = echo "$(shell tty -s && tput setaf 4)$(1):$(shell tty -s && tput sgr0) $(2)"
mesg_step = echo "$(1)"
mesg_ok = echo "result: $(shell tty -s && tput setaf 2)ok$(shell tty -s && tput sgr0)"
mesg_fail = (echo "result: $(shell tty -s && tput setaf 1)fail$(shell tty -s && tput sgr0)" && false)

path_search = $(firstword $(wildcard $(addsuffix /$(1),$(subst :, ,$(PATH)))))

npm_install = \
	$(call mesg_start,main,Installing $(1) via npm...); \
	$(NPM) install $(1) >/dev/null 2>&1 && \
		$(call mesg_ok) || $(call mesg_fail)

TAR ?= tar

GOLINT ?= golint
GOLINT_ARGS =

NPM ?= npm
PATH := $(PATH):$(shell $(NPM) bin)

PANDOC ?= pandoc
PANDOC_ARGS = --standalone --to man

UGLIFYJS ?= uglifyjs
UGLIFYSCRIPT_ARGS = --comments --compress --mangle --screw-ie8
NPM_UGLIFYJS = uglify-js

JSHINT ?= jshint
JSHINT_ARGS =
NPM_JSHINT = jshint

LESSC ?= lessc
LESSC_ARGS = --no-color --clean-css
NPM_LESSC = less
NPM_LESSC_PLUGIN_CLEANCSS = less-plugin-clean-css

all: build

# npm scripts
lessc:
	@if [ -z "$(call path_search,$(LESSC))" ]; then \
		$(call npm_install,$(NPM_LESSC)); \
		$(call npm_install,$(NPM_LESSC_PLUGIN_CLEANCSS)); \
	fi

uglifyjs:
	@if [ -z "$(call path_search,$(UGLIFYJS))" ]; then \
		$(call npm_install,$(NPM_UGLIFYJS)); \
	fi

jshint:
	@if [ -z "$(call path_search,$(JSHINT))" ]; then \
		$(call npm_install,$(NPM_JSHINT)); \
	fi

clean: clean-bin clean-doc clean-static clean-test clean-dist
	@$(call mesg_start,clean,Cleaning source symlink...)
	@rm -rf $(BUILD_DIR)/src && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,clean,Removing build directory...)
	@(test ! -d $(BUILD_DIR) || rmdir $(BUILD_DIR)) && \
		$(call mesg_ok) || $(call mesg_fail)

build: build-bin build-doc build-static

.PHONY: install
install: install-bin install-doc install-static

devel: build devel-static

lint: lint-bin lint-static

test: clean-test test-pkg test-server

$(BUILD_DIR)/src/github.com/facette/facette:
	@$(call mesg_start,main,Creating source symlink...)
	@mkdir -p $(BUILD_DIR)/src/github.com/facette && \
		ln -s ../../../../.. $(BUILD_DIR)/src/github.com/facette/facette && \
		$(call mesg_ok) || $(call mesg_fail)

# Binaries
BIN_SRC = $(wildcard cmd/*/*.go)

BIN_OUTPUT = $(addprefix $(BUILD_DIR)/bin/, $(notdir $(wildcard cmd/*)))

PKG_SRC = $(wildcard pkg/*/*.go)

PKG_LIST = $(wildcard pkg/*)

$(BIN_OUTPUT): $(PKG_SRC) $(BIN_SRC) $(BUILD_DIR)/src/github.com/facette/facette
	@$(call mesg_start,$(notdir $@),Building $(notdir $@)...)
	@install -d -m 0755 $(dir $@) && $(GO) build \
			-ldflags " \
				-X main.version $(VERSION) \
				-X main.buildDate '$(BUILD_DATE)' \
				$(PKG_LIST:%=-X github.com/facette/facette/%.version $(VERSION)) \
				$(PKG_LIST:%=-X github.com/facette/facette/%.buildDate '$(BUILD_DATE)') \
			" \
			-tags "$(TAGS)" \
			-o $@ cmd/$(notdir $@)/*.go && \
		$(call mesg_ok) || $(call mesg_fail)

clean-bin:
	@$(call mesg_start,clean,Cleaning binaries...)
	@rm -rf $(BUILD_DIR)/bin && \
		$(call mesg_ok) || $(call mesg_fail)

build-bin: $(BIN_OUTPUT)

.PHONY: install-bin
install-bin: build-bin
	@$(call mesg_start,install,Installing binaries...)
	@install -d -m 0755 $(PREFIX)/bin && cp $(BIN_OUTPUT) $(PREFIX)/bin && \
		$(call mesg_ok) || $(call mesg_fail)

lint-bin: build-bin
	@$(call mesg_start,lint,Checking sources with Golint...)
	@$(GOLINT) $(GOLINT_ARGS) cmd/... && $(GOLINT) $(GOLINT_ARGS) pkg/... && $(call mesg_ok) || $(call mesg_fail)

# Documentation
MAN_SRC = $(wildcard docs/man/*.[0-9].md)

MAN_OUTPUT = $(addprefix $(BUILD_DIR)/man/, $(notdir $(MAN_SRC:.md=)))

$(MAN_OUTPUT): $(MAN_SRC)
	@$(call mesg_start,docs,Generating $(notdir $@) manual page...)
	@install -d -m 0755 $(dir $@) && $(PANDOC) $(PANDOC_ARGS) docs/man/$(notdir $@).md >$@ && \
		$(call mesg_ok) || $(call mesg_fail)

clean-doc:
	@$(call mesg_start,clean,Cleaning manuals files...)
	@rm -rf $(BUILD_DIR)/man && \
		$(call mesg_ok) || $(call mesg_fail)

build-doc: $(MAN_OUTPUT)

.PHONY: install-doc
install-doc: build-doc
	@$(call mesg_start,install,Installing manuals files...)
	@install -d -m 0755 $(PREFIX)/share && cp -Rp $(BUILD_DIR)/man $(PREFIX)/share && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,install,Installing examples files...)
	@install -d -m 0755 $(PREFIX)/share/facette/examples && cp -Rp docs/examples $(PREFIX)/share/facette && \
		$(call mesg_ok) || $(call mesg_fail)

# Static
SCRIPT_SRC = cmd/facette/js/intro.js \
	cmd/facette/js/define.js \
	cmd/facette/js/extend.js \
	cmd/facette/js/ajax.js \
	cmd/facette/js/utils.js \
	cmd/facette/js/setup.js \
	cmd/facette/js/console.js \
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
	cmd/facette/js/admin/collection.js \
	cmd/facette/js/admin/graph.js \
	cmd/facette/js/admin/group.js \
	cmd/facette/js/admin/scale.js \
	cmd/facette/js/admin/unit.js \
	cmd/facette/js/admin/catalog.js \
	cmd/facette/js/admin/outro.js \
	cmd/facette/js/browse/intro.js \
	cmd/facette/js/browse/browse.js \
	cmd/facette/js/browse/collection.js \
	cmd/facette/js/browse/graph.js \
	cmd/facette/js/browse/outro.js \
	cmd/facette/js/item.js \
	cmd/facette/js/graph.js \
	cmd/facette/js/pane.js \
	cmd/facette/js/i18n.js \
	cmd/facette/js/outro.js

SCRIPT_OUTPUT = $(BUILD_DIR)/static/facette.js

SCRIPT_EXTRA = cmd/facette/js/thirdparty/jquery.js \
	cmd/facette/js/thirdparty/jquery.datepicker.js \
	cmd/facette/js/thirdparty/highcharts.js \
	cmd/facette/js/thirdparty/highcharts.exporting.js \
	cmd/facette/js/thirdparty/i18next.js \
	cmd/facette/js/thirdparty/moment.js \
	cmd/facette/js/thirdparty/canvg.js \
	cmd/facette/js/thirdparty/rgbcolor.js \
	cmd/facette/js/thirdparty/sprintf.js

SCRIPT_EXTRA_OUTPUT = $(addprefix $(BUILD_DIR)/static/, $(notdir $(SCRIPT_EXTRA)))

MESG_SRC = cmd/facette/js/messages.json

MESG_OUTPUT = $(BUILD_DIR)/static/$(notdir $(MESG_SRC))

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

STYLE_OUTPUT = $(BUILD_DIR)/static/style.css

STYLE_PRINT_SRC = cmd/facette/style/intro.less \
	cmd/facette/style/define.less \
	cmd/facette/style/common-print.less

STYLE_PRINT_OUTPUT = $(BUILD_DIR)/static/style.print.css

STYLE_EXTRA = cmd/facette/style/extra/favicon.png \
	cmd/facette/style/extra/fonts \
	cmd/facette/style/extra/loader.gif \
	cmd/facette/style/extra/logo-text.png \
	cmd/facette/style/extra/logo-text-light.png

STYLE_EXTRA_OUTPUT = $(addprefix $(BUILD_DIR)/static/, $(notdir $(STYLE_EXTRA)))

TMPL_SRC = $(wildcard cmd/facette/template/*/*.html) \
	$(wildcard cmd/facette/template/*.html) \
	$(wildcard cmd/facette/template/*.xml)

TMPL_OUTPUT = $(BUILD_DIR)/template

$(SCRIPT_OUTPUT): uglifyjs $(SCRIPT_SRC)
	@$(call mesg_start,static,Merging script files into $(notdir $(SCRIPT_OUTPUT:.js=.src.js))...)
	@install -d -m 0755 $(BUILD_DIR)/static && cat $(SCRIPT_SRC) >$(SCRIPT_OUTPUT:.js=.src.js) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,static,Packing $(notdir $(SCRIPT_OUTPUT:.js=.src.js)) file...)
	@$(UGLIFYJS) $(UGLIFYSCRIPT_ARGS) --output $(SCRIPT_OUTPUT) $(SCRIPT_OUTPUT:.js=.src.js) && \
		$(call mesg_ok) || $(call mesg_fail)

$(SCRIPT_EXTRA_OUTPUT): $(SCRIPT_EXTRA)
	@$(call mesg_start,static,Copying third-party files...)
	@cp -r $(SCRIPT_EXTRA) $(BUILD_DIR)/static/ && \
		$(call mesg_ok) || $(call mesg_fail)

$(MESG_OUTPUT): $(MESG_SRC)
	@$(call mesg_start,static,Packing $(MESG_SRC) file...)
	@install -d -m 0755 $(BUILD_DIR)/static && \
		sed -e 's/^\s\+//g;s/\s\+$$//g' $(MESG_SRC) | sed -e ':a;N;s/\n//;ta' >$(MESG_OUTPUT) && \
		$(call mesg_ok) || $(call mesg_fail)

$(STYLE_OUTPUT): lessc $(STYLE_SRC)
	@$(call mesg_start,static,Merging style files into $(notdir $(STYLE_OUTPUT:.css=.src.css))...)
	@install -d -m 0755 $(BUILD_DIR)/static && cat $(STYLE_SRC) >$(STYLE_OUTPUT:.css=.src.css) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,static,Packing $(notdir $(STYLE_OUTPUT:.css=.src.css)) file...)
	@$(LESSC) $(LESSC_ARGS) $(STYLE_OUTPUT:.css=.src.css) >$(STYLE_OUTPUT) && \
		$(call mesg_ok) || $(call mesg_fail)

$(STYLE_PRINT_OUTPUT): lessc $(STYLE_PRINT_SRC)
	@$(call mesg_start,static,Merging style files into $(notdir $(STYLE_PRINT_OUTPUT:.css=.src.css))...)
	@install -d -m 0755 $(BUILD_DIR)/static && cat $(STYLE_PRINT_SRC) >$(STYLE_PRINT_OUTPUT:.css=.src.css) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,static,Packing $(notdir $(STYLE_PRINT_OUTPUT:.css=.src.css)) file...)
	@$(LESSC) $(LESSC_ARGS) $(STYLE_PRINT_OUTPUT:.css=.src.css) >$(STYLE_PRINT_OUTPUT) && \
		$(call mesg_ok) || $(call mesg_fail)

$(STYLE_EXTRA_OUTPUT): $(STYLE_EXTRA)
	@$(call mesg_start,static,Copying extra files...)
	@cp -r $(STYLE_EXTRA) $(BUILD_DIR)/static/ && \
		$(call mesg_ok) || $(call mesg_fail)

$(TMPL_OUTPUT): $(TMPL_SRC)
ifeq ($(UNAME), Darwin)
	$(eval COPY_CMD=rsync -rR)
else
	$(eval COPY_CMD=cp -r --parents)
endif
	@$(call mesg_start,build,Copying template files...)
	@install -d -m 0755 $(BUILD_DIR)/template && \
		(cd cmd/facette/template; $(COPY_CMD) $(TMPL_SRC:cmd/facette/template/%=%) ../../../$(TMPL_OUTPUT)) && \
		$(call mesg_ok) || $(call mesg_fail)

clean-static:
	@$(call mesg_start,clean,Cleaning static files...)
	@rm -rf $(BUILD_DIR)/static $(BUILD_DIR)/template && \
		$(call mesg_ok) || $(call mesg_fail)

build-static: $(SCRIPT_OUTPUT) $(SCRIPT_EXTRA_OUTPUT) $(MESG_OUTPUT) $(STYLE_OUTPUT) $(STYLE_PRINT_OUTPUT) \
	$(STYLE_EXTRA_OUTPUT) $(TMPL_OUTPUT)

.PHONY: install-static
install-static: build-static $(TMPL_SRC)
	@$(call mesg_start,install,Installing static files...)
	@install -d -m 0755 $(PREFIX)/share/facette/static && cp -Rp $(SCRIPT_OUTPUT) $(SCRIPT_EXTRA_OUTPUT) \
		$(MESG_OUTPUT) $(STYLE_OUTPUT) $(STYLE_PRINT_OUTPUT) $(STYLE_EXTRA_OUTPUT) $(PREFIX)/share/facette/static && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,install,Installing template files...)
	@cp -Rp $(TMPL_OUTPUT) $(PREFIX)/share/facette && \
		$(call mesg_ok) || $(call mesg_fail)

.PHONY: devel-static
devel-static: build-static
	@$(call mesg_start,install,Copying static development files...)
	@cp $(SCRIPT_OUTPUT:.js=.src.js) $(BUILD_DIR)/static/$(notdir $(SCRIPT_OUTPUT)) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,install,Copying static third-party development files...)
	@(for ENTRY in $(SCRIPT_EXTRA:.js=.src.js); do \
		cp $$ENTRY $(BUILD_DIR)/static/`basename $$ENTRY | sed -e 's@\.src\.js$$@.js@'`; \
	done) && $(call mesg_ok) || $(call mesg_fail)

lint-static: jshint $(SCRIPT_OUTPUT)
	@$(call mesg_start,lint,Checking $(notdir $(SCRIPT_OUTPUT:.js=.src.js)) with JSHint...)
	-@$(JSHINT) $(JSHINT_ARGS) $(SCRIPT_OUTPUT:.js=.src.js) && \
		$(call mesg_ok) || $(call mesg_fail)

# Test
TEST_DIR = $(BUILD_DIR)/tests

$(TEST_DIR):
	@install -d -m 0755 $(TEST_DIR)

$(PKG_LIST): $(TEST_DIR) $(BUILD_DIR)/src/github.com/facette/facette
	@$(call mesg_start,test,Testing $@ package...)
	@(cd $(TEST_DIR) && $(GO) test -race -c -i ../../../$@ && \
		(test ! -f ./$(@:pkg/%=%).test || ./$(@:pkg/%=%).test -test.v=true) && \
		$(call mesg_ok) || $(call mesg_fail))

clean-test:
	@$(call mesg_start,clean,Cleaning test data...)
	@rm -rf $(BUILD_DIR)/tests $(BUILD_DIR)/pkg && \
		$(call mesg_ok) || $(call mesg_fail)

test-pkg: $(PKG_LIST)

test-server: $(TEST_DIR) build-bin
	@$(call mesg_start,test,Starting facette server...)
	@(cd $(BUILD_DIR); bin/facette -c ../../tests/facette.json -l tests/facette.log -L debug &) && \
		$(call mesg_ok) || $(call mesg_fail)

	@$(call mesg_start,test,Running server tests...)
	@(cd $(BUILD_DIR)/tests; $(GO) test -race -c -i ../../../cmd/facette) && \
		(cd $(BUILD_DIR)/; ./tests/facette.test -test.v=true -c ../../tests/facette.json) && \
		$(call mesg_ok) || (kill -2 `cat $(BUILD_DIR)/tests/facette.pid`; $(call mesg_fail))

	@$(call mesg_start,test,Stopping facette server...)
	@kill -2 `cat $(BUILD_DIR)/tests/facette.pid` && \
		$(call mesg_ok) || $(call mesg_fail)

# Distribution
DIST_DIR = dist
DIST_BUILD_DIR = $(DIST_DIR)/$(BUILD_NAME)

clean-dist:
	@$(call mesg_start,clean,Cleaning distribution files...)
	@rm -rf $(DIST_DIR)/facette-* && \
		$(call mesg_ok) || $(call mesg_fail)

.PHONY: dist
dist:
	@$(call mesg_start,dist,Creating distribution disrectory...)
	@install -d -m 0755 $(DIST_DIR)/$(BUILD_NAME) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(MAKE) PREFIX=$(DIST_DIR)/$(BUILD_NAME) --no-print-directory install
	@$(call mesg_start,dist,Build distribution tarball...)
	@install -d -m 0755 $(DIST_DIR) && \
		$(TAR) -C $(DIST_DIR) -czf $(DIST_DIR)/$(BUILD_NAME:facette-%=facette-$(VERSION)-%).tar.gz $(BUILD_NAME) && \
		$(call mesg_ok) || $(call mesg_fail)

DOCKER_TAG ?= facette-latest

.PHONY: dist
docker:
	docker build -t $(DOCKER_TAG) .
