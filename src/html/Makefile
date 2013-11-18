#!/usr/bin/make -f
# -*- Makefile -*-

include ../../Makefile.inc

DST_DIR = $(PREFIX)/share/html

SRC_FILES = common \
	admin \
	browse \
	wait.html \
	layout.html

all: install

clean:
	@$(call mesg_start,Cleaning HTML files...)
	@rm -rf $(DST_DIR) && $(call mesg_ok) || $(call mesg_fail)
	@$(call clean_build)

lint:
	@#noop

build: $(SRC_FILES)

test:
	@#noop

install: build
	@$(call mesg_start,Copying HTML files...)
	@install -m 0755 -d $(DST_DIR) && cp -r $(SRC_FILES) $(DST_DIR)/ && $(call mesg_ok) || $(call mesg_fail)

devel: install
