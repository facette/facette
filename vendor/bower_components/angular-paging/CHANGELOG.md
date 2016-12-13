####v2.2.2: (2016-06-04) Adding Disabled Attribute

New Features
- Added Disabled attribute support 
- How to use section added to the readme

Updates
- Limited minimum Angular version to 1.3.1 due to scope watch support
- Updated readme code sample section

---
<br/>

####v2.2.1: (2016-2-18) Bower hotfix 

Updates
- Corrected bad Angular 1.5.0 reference in bower.json 

---
<br/>

####v2.2.0: (2016-2-12) CSS Class Additions and Attribute Cleanup

New Features
- Added support for custom CSS first, last, next, previous text like glyphicons in bootstrap
- Added Angular dependency in bower
- Adjusted boolean logic for flag settings like scroll-top

Updates
- Added tests to support new features
- Updated to Angular 1.5.0
- Updated to PhantomJS2 for tests

---
<br/>

####v2.1.0: (2015-12-22) Localization and Link Support

New Features
- Added support for custom first, last, next, previous text
- Added granularity to show combinations of first, last, next, previous items
- Added disabled class to the dots list item 
- Switched from span tags to anchor tags in list items
- Introduced {page} text format to display the page number
- Introduced {page} href format to display the page number 

Updates
- Added tests to support new features
- Split and created individual 'spec' files for testing  
- Added specific advanced test to index.html
- Updated to bootstrap 3.3.6 
- Updated to angular 1.4.8

---
<br/>

####v2.0.0: (2015-08-30) Node and Structure Improvements

Breaking Changes
- Changed module name from `brantwills.paging` to `bw.paging`
- File structure changed to support distribution (dist) and source (src) folders

New Features
- Added travis-ci support
- Added grunt support
- Added npm package support

Updates
- Changed to phantomJs for npm testing
- Moved `paging.js` functions around bringing most important info to the top

---
<br/>

####v1.0.3: (2015-05-07) Testing and Improvements 

- Added Karma suite with Jasmine tests
- Improved adjacent logic for beginning and end displays
- Improved documentation in code
- Combined previous next function logic - shorten code

---
<br/>

####v1.0.2: (2015-04-23) Code Enhancements 

- Added internal paging actions for pageSize and total directive attributes
- Added pageSize to watch
- Fixed tabs to space formatting
- Updated Angular and Bootstrap

---
<br/>

####v1.0.1: (2015-03-02) Bug fixes and resource updates

- Minor bug fixes around the dot logic
- Rearranged some of the functions 
- Updated angular resources
- Improved module naming convention

---
<br/>

####v1.0.0: (2014-12-28) First Release to bower

- First release to support Bower
- Requires at minimum AngularJS 1.2.x
