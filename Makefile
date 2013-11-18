#!/usr/bin/make -f
# -*- Makefile -*-

SUB_DIRS=src

all %:
	@for i in $(SUB_DIRS); do make --no-print-directory -C $$i $@; done
