/*!
 * Facette - Web graphing front-end
 * @author   Vincent Batoufflet <vincent@facette.io>
 * @link     https://facette.io/
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

var $head = $(document.head),
    $body = $(document.body),
    $window = $(window);

// Get location path
var locationPath = String(window.location.pathname);

// Get URL prefix
var urlPrefix = $head.find('meta[name=url-prefix]').attr('content') || '',
    readOnly = $head.find('meta[name=read-only]').attr('content') == 'true';
