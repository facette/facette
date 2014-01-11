#!/usr/bin/make -f
# -*- Makefile -*-

include ../../Makefile.inc

PANDOC ?= pandoc
PANDOC_ARGS = --standalone --to man

DST_DIR = $(PREFIX)/share/man

SRC_FILES = $(wildcard *.[0-9].md)

all: install

clean:
	@$(call mesg_start,Cleaning manpages files...)
	@rm -rf $(DST_DIR) && $(call mesg_ok) || $(call mesg_fail)
	@$(call mesg_start,Cleaning temporary files...)
	@(test ! -d $(TEMP_DIR) || (cd $(TEMP_DIR) && rm -rf $(basename $(SRC_FILES)))) && \
		$(call mesg_ok) || $(call mesg_fail)
	@$(call clean_build)

lint:
	@#noop

build: $(SRC_FILES)
	@$(call mesg_start,Generating manpages...)
	@for ENTRY in $(SRC_FILES); do \
		$(PANDOC) $(PANDOC_ARGS) $$ENTRY >$(TEMP_DIR)/`echo $$ENTRY | sed -e 's/\.md$$//'`; \
		done && \
		$(call mesg_ok) || $(call mesg_fail)

test:
	@#noop

install: build
	@$(call mesg_start,Copying manpages files...)
	@install -m 0755 -d $(DST_DIR) && (cd $(TEMP_DIR) && cp -r $(basename $(SRC_FILES)) $(DST_DIR)/) && \
		$(call mesg_ok) || $(call mesg_fail)

devel: install
