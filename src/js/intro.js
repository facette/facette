/*!
 * Facette - Web graphing front-end
 * @author   Vincent Batoufflet <vincent@facette.io>
 * @link     http://facette.io/
 * @license  BSD
 */

$(function () {

/*jshint
    browser: true,
    devel: true,
    jquery: true,
    trailing: true
 */

/*globals
    canvg,
    Highcharts,
    moment
 */

"use strict";

var $body = $(document.body),
    $window = $(window);

// Get URL prefix
var urlPrefix = $(document.head).find('meta[name=url-prefix]').attr('content') || '';
