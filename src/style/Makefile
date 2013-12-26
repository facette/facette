#!/usr/bin/make -f
# -*- Makefile -*-

include ../../Makefile.inc

LESSC = lessc
LESSC_ARGS = --no-color

CSS_FILE = $(TEMP_DIR)/style.css
CSS_FILE_MIN = $(TEMP_DIR)/style.min.css

CSS_FILE_PRINT = $(TEMP_DIR)/style.print.css
CSS_FILE_PRINT_MIN = $(TEMP_DIR)/style.print.min.css

DST_DIR = $(PREFIX)/share/static

SRC_FILES = intro.less \
	define.less \
	font.less \
	common.less \
	icon.less \
	form.less \
	tooltip.less \
	date.less \
	overlay.less \
	list.less \
	stats.less \
	admin.less \
	graph.less

SRC_FILES_PRINT = intro.less \
	define.less \
	common-print.less

EXT_FILES = fonts \
	favicon.png \
	logo.png \
	loader.gif

all: install

clean:
	@$(call mesg_start,Cleaning installed files...)
	@(test ! -d $(DST_DIR) || (cd $(DST_DIR) && rm -rf $(notdir $(CSS_FILE) $(CSS_FILE_MIN) $(CSS_FILE_PRINT) \
		$(CSS_FILE_PRINT_MIN)) $(EXT_FILES))) && $(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,Cleaning temporary files...)
	@rm -f $(CSS_FILE) $(CSS_FILE_MIN) $(CSS_FILE_PRINT) $(CSS_FILE_PRINT_MIN) && $(call mesg_ok) || $(call mesg_fail)
	@$(call clean_build)

$(CSS_FILE): $(TEMP_DIR) $(SRC_FILES)
	@$(call mesg_start,Merging source files into $(notdir $(CSS_FILE))...)
	@cat $(SRC_FILES) | $(LESSC) $(LESSC_ARGS) - >$(CSS_FILE) && $(call mesg_ok) || $(call mesg_fail)

$(CSS_FILE_MIN): $(TEMP_DIR) $(CSS_FILE)
	@$(call mesg_start,Packing $(notdir $(CSS_FILE)) file...)
	@$(LESSC) $(LESSC_ARGS) --yui-compress $(CSS_FILE) >$(CSS_FILE_MIN) && $(call mesg_ok) || $(call mesg_fail)

$(CSS_FILE_PRINT): $(TEMP_DIR) $(SRC_FILES_PRINT)
	@$(call mesg_start,Merging source files into $(notdir $(CSS_FILE_PRINT))...)
	@cat $(SRC_FILES_PRINT) | $(LESSC) $(LESSC_ARGS) - >$(CSS_FILE_PRINT) && $(call mesg_ok) || $(call mesg_fail)

$(CSS_FILE_PRINT_MIN): $(TEMP_DIR) $(CSS_FILE_PRINT)
	@$(call mesg_start,Packing $(notdir $(CSS_FILE_PRINT)) file...)
	@$(LESSC) $(LESSC_ARGS) --yui-compress $(CSS_FILE_PRINT) >$(CSS_FILE_PRINT_MIN) && \
		$(call mesg_ok) || $(call mesg_fail)

lint:
	@#noop

build: $(CSS_FILE_MIN) $(CSS_FILE_PRINT_MIN)

test:
	@#noop

install: build
	@$(call mesg_start,Copying built files...)
	@install -m 0755 -d $(DST_DIR) && \
		install -m 0644 $(CSS_FILE_MIN) $(DST_DIR)/$(notdir $(CSS_FILE)) && \
		install -m 0644 $(CSS_FILE_PRINT_MIN) $(DST_DIR)/$(notdir $(CSS_FILE_PRINT)) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,Copying extra files...)
	@cd extra && cp -r $(EXT_FILES) $(DST_DIR)/ && $(call mesg_ok) || $(call mesg_fail)

devel: install
	@$(call mesg_start,Overriding files...)
	@install -m 0644 $(CSS_FILE) $(DST_DIR)/$(notdir $(CSS_FILE)) && \
		install -m 0644 $(CSS_FILE_PRINT) $(DST_DIR)/$(notdir $(CSS_FILE_PRINT)) && \
		$(call mesg_ok) || $(call mesg_fail)
