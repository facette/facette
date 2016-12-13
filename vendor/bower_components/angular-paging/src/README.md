####Source Files
The source file directory contains a sample working application using all the features of the paging directive.

The goal of this working version is to mitigate simple setup and "how to" questions as well as visually test new features.

---
<br/>

####Paging.js

The angular paging directive javascript code.  

This is the latest code under development - see the `dist` folder for tagged and minified versions.

The paging directive is contained in a `bw.paging` module which you can consume in your project.

<br/>
**The following constraints are built into the directive by design:**

- If the current page value is larger than the page count, the page will be set to the page count value

- If the current page value is less than or equal to zero (0), the page will be set to one (1)

- If the adjacent count value is less than or equal to zero (0), the adjacent count value will be set to two (2)

- If the page size does not exist or is less than or equal to zero (0), the page size is set to one (1)

- The current page on click event is disabled

---
<br/>

####App.js

A simple angular application module which consumes the paging directive and introduces a single controller. 

The controller is used to explain how a `paging action` could be introduced into the paging directive.

---
<br/>

####Index.html

A simple HTML implementation of the angular application defined in `app.js`

This file should exercise both the simple and advanced options of the paging directive.

For simplicity we are using [bootstrap](http://getbootstrap.com/) for styling. 
