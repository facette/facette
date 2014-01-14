#!/usr/bin/make -f
# -*- Makefile -*-

BASE_DIR = $(realpath $(dir $(lastword $(MAKEFILE_LIST))))

BUILD_DIR = $(BASE_DIR)/build
TEMP_DIR = $(BUILD_DIR)/tmp

MAKEFILE_DIR = $(realpath $(dir $(firstword $(MAKEFILE_LIST))))

PREFIX ?= $(BUILD_DIR)/facette

mesg_start = echo "$(shell tput setaf 4)$(shell basename $(CURDIR)):$(shell tput sgr0) $(1)"
mesg_step = echo "$(1)"
mesg_ok = echo "result: $(shell tput setaf 2)ok$(shell tput sgr0)"
mesg_fail = (echo "result: $(shell tput setaf 1)fail$(shell tput sgr0)" && false)

clean_build = test ! -d $(BUILD_DIR) || find $(BUILD_DIR) -type d -empty -delete

$(TEMP_DIR):
	@mkdir -p $(TEMP_DIR)
