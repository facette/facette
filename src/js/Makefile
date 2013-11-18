#!/usr/bin/make -f
# -*- Makefile -*-

include ../../Makefile.inc

UGLIFYJS = uglifyjs
UGLIFYJS_ARGS = --comments --compress --mangle --screw-ie8

JSHINT = jshint
JSHINT_ARGS = --show-non-errors

JS_FILE = $(TEMP_DIR)/facette.js
JS_FILE_MIN = $(TEMP_DIR)/facette.min.js

MSG_FILE = messages.json
MSG_FILE_MIN = $(TEMP_DIR)/messages.min.json

DST_DIR = $(PREFIX)/share/static

SRC_FILES = intro.js \
	define.js \
	extend.js \
	utils.js \
	setup.js \
	overlay.js \
	list.js \
	tree.js \
	menu.js \
	input.js \
	select.js \
	link.js \
	admin/intro.js \
	admin/admin.js \
	admin/graph.js \
	admin/collection.js \
	admin/group.js \
	admin/outro.js \
	browse/intro.js \
	browse/browse.js \
	browse/collection.js \
	browse/outro.js \
	graph.js \
	collection.js \
	group.js \
	stats.js \
	pane.js \
	i18n.js \
	outro.js

EXT_FILES_MIN = jquery.min.js \
	jquery.datepicker.min.js \
	i18next.min.js \
	highcharts.min.js \
	highcharts.exporting.min.js \
	rgbcolor.min.js \
	canvg.min.js \
	moment.min.js

EXT_FILES = $(EXT_FILES_MIN:.min.js=.js)

all: lint install

clean:
	@$(call mesg_start,Cleaning installed files...)
	@(test ! -d $(DST_DIR) || (cd $(DST_DIR) && rm -rf $(notdir $(JS_FILE) $(MSG_FILE)) $(EXT_FILES))) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,Cleaning temporary files...)
	@rm -f $(JS_FILE) $(JS_FILE_MIN) $(MSG_FILE_MIN) && $(call mesg_ok) || $(call mesg_fail)
	@$(call clean_build)

$(JS_FILE): $(TEMP_DIR) $(SRC_FILES)
	@$(call mesg_start,Merging origin files into $(JS_FILE)...)
	@cat $(SRC_FILES) >$(JS_FILE) && $(call mesg_ok) || $(call mesg_fail)

$(JS_FILE_MIN): $(TEMP_DIR) $(JS_FILE)
	@$(call mesg_start,Packing $(JS_FILE) file...)
	@$(UGLIFYJS) $(UGLIFYJS_ARGS) --output $(JS_FILE_MIN) $(JS_FILE) && $(call mesg_ok) || $(call mesg_fail)

$(MSG_FILE_MIN): $(TEMP_DIR) $(MSG_FILE)
	@$(call mesg_start,Packing $(MSG_FILE) file...)
	@sed -e 's/^\s\+//g;s/\s\+$$//g' $(MSG_FILE) | sed -e ':a;N;s/\n//;ta' >$(MSG_FILE_MIN) && \
		$(call mesg_ok) || $(call mesg_fail)

lint: $(JS_FILE)
	@$(call mesg_start,Checking $(JS_FILE) with JSHint...)
	-@$(JSHINT) $(JS_FILE) $(JSHINT_ARGS) && $(call mesg_ok) || $(call mesg_fail)

build: $(JS_FILE_MIN) $(MSG_FILE_MIN)

test:
	@#noop

install: build
	@$(call mesg_start,Copying built files...)
	@install -m 0755 -d $(DST_DIR) && \
		install -m 0644 $(JS_FILE_MIN) $(DST_DIR)/$(notdir $(JS_FILE)) && \
		install -m 0644 $(MSG_FILE_MIN) $(DST_DIR)/$(notdir $(MSG_FILE)) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,Copying extra files...)
	@cd extra && for FILE in $(EXT_FILES_MIN); do \
		cp -r $$FILE $(DST_DIR)/`echo $$FILE | sed -e 's/\.min\.js$$/.js/'`; \
		done && $(call mesg_ok) || $(call mesg_fail)

devel: install
	@$(call mesg_start,Overriding files...)
	@install -m 0644 $(JS_FILE) $(DST_DIR)/$(notdir $(JS_FILE)) && \
		install -m 0644 $(MSG_FILE) $(DST_DIR)/$(notdir $(MSG_FILE)) && \
		cd extra && cp -r $(EXT_FILES) $(DST_DIR)/ && \
		$(call mesg_ok) || $(call mesg_fail)
