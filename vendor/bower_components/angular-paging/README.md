# Angular Paging
[![npm version](https://img.shields.io/npm/v/angular-paging.svg)](https://www.npmjs.org/package/angular-paging)
[![bower version](https://img.shields.io/bower/v/angular-paging.svg)](https://www.npmjs.org/package/angular-paging)
[![Build Status](https://travis-ci.org/brantwills/Angular-Paging.svg)](https://travis-ci.org/brantwills/Angular-Paging)
[![CDN](https://img.shields.io/badge/cdn-rawgit-brightgreen.svg)](https://rawgit.com/brantwills/Angular-Paging/master/dist/paging.min.js) 


**Demo Available At: [http://brantwills.github.io/Angular-Paging/](http://brantwills.github.io/Angular-Paging/)**


An Angular directive to aid paging large datasets requiring minimum paging information.  This paging directive is unique in that we are only interested in the active page of items rather than holding the entire list of items in memory.  This forces any filtering or sorting to be performed outside the directive.

## Background
I have often found myself paging across millions of log rows or massive non-normalized lists even after 
some level of filtering by date range or on some column value.  These scenarios have pushed me to develop a reusable paging scheme which just happens to drop nicely into AngularJS.

## Installation and Contribution
The core of this project is a simple angular directive which allows you to use the code in many different ways.  If you are interested in keeping current with bug fixes and features, we support both bower and npm install commands.  If you just want to grab the latest or work with CDN's, head over to the distribution folder for the latest code base.  Finally, if you are interested in contributing or see any issues feel free to fork and test away!

## Blah Blah Blah.. How to Use!
To include the paging directive in your own project, add the `paging.js` or `paging.min.js` file and include the module as a dependency to your angular application.  We do support **[npm](https://www.npmjs.org/package/angular-paging)** and **[bower](http://bower.io/)** if you are familiar with those distribution systems.  Please review the **[src/index.html](https://github.com/brantwills/Angular-Paging/blob/master/src/index.html)** and **[src/app.js](https://github.com/brantwills/Angular-Paging/blob/master/src/app.js)** files for a working version of the directive if you are new to angular modules.
``` javascript
// Add the Angular-Paging module as a dependency to your application module:
var app = angular.module('yourApp', ['bw.paging'])
```

<br/>
<br/>

## Code Samples
**See [Full Demo](http://brantwills.github.io/Angular-Paging/) for complete samples and documentation**

The following attributes explored in the basic example are required directive inputs:

1. `page` What page am I currently viewing
2. `pageSize` How many items in the list to display on a page
3. `total` What is the total count of items in my list

The other code examples explore supporting attributes which may be mixed and matched as you see fit. Please see **[src/index.html](https://github.com/brantwills/Angular-Paging/blob/master/src/index.html)** for complete code samples and documentation for a working HTML sample.

<br/>

**Basic Example**

[![alt text](https://raw.githubusercontent.com/brantwills/Angular-Paging/gh-pages/basicSample.png "Basic Sample")](http://brantwills.github.io/Angular-Paging/)
```html
<div paging
  page="35" 
  page-size="10" 
  total="1000"
  paging-action="foo('bar', page)">
</div> 
```

<br/>

**Enable First and Last Text**

[![alt text](https://raw.githubusercontent.com/brantwills/Angular-Paging/gh-pages/advancedSample.png "Basic Sample")](http://brantwills.github.io/Angular-Paging/)
```html
<paging
  page="currentPage" 
  page-size="pageSize" 
  total="total"
  show-prev-next="true"
  show-first-last="true">
</paging>  
```

<br/>

**Adjust Text, Class, and Hover Over Title** 

```html
<paging
  ...
  text-first="&laquo;"
  text-last="&raquo;"
  text-next="&rsaquo;"
  text-prev="&lsaquo;"
  text-title-page="Page {page} hover title text"
  text-title-first="First Page hover title text"
  text-title-last="Last Page hover title text"
  text-title-next="Next Page hover title text"
  text-title-prev="Previous hover Page title text"  
  text-first-class="glyphicon glyphicon-backward"
  text-last-class="glyphicon glyphicon-forward" 
  text-next-class="glyphicon glyphicon-chevron-right"
  text-prev-class="glyphicon glyphicon-chevron-left">
</paging>  
```

<br/>

**Enable Anchor Link Href**

The text `{page}` will display the page number
```html
<paging
  ...
  pg-href="#GotoPage-{page}">
</paging>   
```

<br/>

**Adjust Class Name Settings**

```html
<paging
  ...
  class="small"
  ul-class="{{ulClass}}"
  active-class="{{activeClass}}"
  disabled-class="{{disabledClass}}">
</paging>   
```

<br/>

**Boolean Flag Settings**

```html
<paging
  ...
  disabled="{{isDisabled}}"
  scroll-top="{{willScrollTop}}" 
  hide-if-empty="{{hideIfEmpty}}">
</paging>   
```

<br/>

**Other Helper Settings**

```html
<paging
  ...
  adjacent="{{adjacent}}"
  dots="{{dots}}"
  paging-action="DoCtrlPagingAct('Paging Clicked', page, pageSize, total)">
</paging>   


