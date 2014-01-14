#!/usr/bin/make -f
# -*- Makefile -*-

SUB_DIRS=$(wildcard */)

all %:
	@for i in $(SUB_DIRS); do make --no-print-directory -C $$i $@ || exit $$?; done
