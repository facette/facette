#!/usr/bin/make -f
# -*- Makefile -*-

SED = sed

CURL = curl
CURL_ARGS = --silent

UGLIFYJS = uglifyjs
UGLIFYJS_ARGS = --comments --compress --mangle --screw-ie8

all:
	@echo "Use \`make update' to update third-party libraries."

update:
	@echo 'Updating Highcharts...'
	@$(CURL) $(CURL_ARGS) --output highcharts.js \
		http://code.highcharts.com/highcharts.src.js
	@$(CURL) $(CURL_ARGS) --output highcharts.min.js \
		http://code.highcharts.com/highcharts.js
	@$(CURL) $(CURL_ARGS) --output highcharts.exporting.js \
		http://code.highcharts.com/modules/exporting.src.js
	@$(CURL) $(CURL_ARGS) --output highcharts.exporting.min.js \
		http://code.highcharts.com/modules/exporting.js
	@echo 'Updating Moment.js...'
	@$(CURL) $(CURL_ARGS) --output moment.js \
		https://raw.github.com/moment/moment/master/moment.js
	@$(CURL) $(CURL_ARGS) --output moment.min.js \
		https://raw.github.com/moment/moment/master/min/moment.min.js
	@echo 'Updating jquery-simple-datetimepicker...'
	@$(CURL) $(CURL_ARGS) --output jquery.datepicker.js \
		https://raw.github.com/mugifly/jquery-simple-datetimepicker/master/jquery.simple-dtpicker.js
	@$(SED) -re 's/\s+$$//' -i jquery.datepicker.js
	@$(UGLIFYJS) $(UGLIFYJS_ARGS) --comments '/jquery-simple-datetimepicker/' -- \
		jquery.datepicker.js >jquery.datepicker.min.js
	@echo 'Updating canvg...'
	@$(CURL) $(CURL_ARGS) --output canvg.js \
		http://canvg.googlecode.com/svn/trunk/canvg.js
	@$(UGLIFYJS) $(UGLIFYJS_ARGS) --comments '/canvg.js/' -- canvg.js >canvg.min.js
	@$(CURL) $(CURL_ARGS) --output rgbcolor.js \
		http://canvg.googlecode.com/svn/trunk/rgbcolor.js
	@$(UGLIFYJS) $(UGLIFYJS_ARGS) -- rgbcolor.js >rgbcolor.min.js
