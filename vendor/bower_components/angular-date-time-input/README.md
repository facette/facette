# Angular Date/Time input
================================

Native AngularJS directive that allows user input of a date/time value. Valid dates are displayed in specified format, but input may be in any supported format.

[![Join the chat at https://gitter.im/dalelotts/angular-date-time-input](https://badges.gitter.im/dalelotts/angular-date-time-input.svg)](https://gitter.im/dalelotts/angular-date-time-input?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![MIT License][license-image]][license-url]
[![Build Status](https://travis-ci.org/dalelotts/angular-date-time-input.png?branch=master)](https://travis-ci.org/dalelotts/angular-date-time-input)
[![Coverage Status](https://coveralls.io/repos/github/dalelotts/angular-date-time-input/badge.svg?branch=master)](https://coveralls.io/github/dalelotts/angular-date-time-input?branch=master)
[![Dependency Status](https://david-dm.org/dalelotts/angular-date-time-input.svg)](https://david-dm.org/dalelotts/angular-date-time-input)
[![devDependency Status](https://david-dm.org/dalelotts/angular-date-time-input/dev-status.svg)](https://david-dm.org/dalelotts/angular-date-time-input#info=devDependencies)
[![JavaScript Style Guide](https://img.shields.io/badge/code%20style-standard-brightgreen.svg)](http://standardjs.com/)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)
[![Commitizen friendly](https://img.shields.io/badge/commitizen-friendly-brightgreen.svg)](http://commitizen.github.io/cz-cli/)
[![PayPal donate button](http://img.shields.io/paypal/donate.png?color=yellow)](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=F3FX5W6S2U4BW&lc=US&item_name=Dale%20Lotts&item_number=angular%2dbootstrap%2ddatetimepicker&currency_code=USD&bn=PP%2dDonationsBF%3abtn_donate_SM%2egif%3aNonHosted "Donate one-time to this project using Paypal")
<a href="https://twitter.com/intent/tweet?original_referer=https%3A%2F%2Fabout.twitter.com%2Fresources%2Fbuttons&amp;text=Check%20out%20this%20%23AngularJS%20directive%20that%20makes%20it%20dead%20simple%20for%20users%20to%input%20dates%20%26%20times&amp;tw_p=tweetbutton&amp;url=https%3A%2F%2Fgithub.com%2Fdalelotts%2Fangular-date-time-input&amp;via=dalelotts" target="_blank">
  <img src="http://jpillora.com/github-twitter-button/img/tweet.png"></img>
</a>

#Dependencies

Requires:
 * AngularJS 1.4.x or higher
 * MomentJS 2.1.x or higher

#Testing
We use karma and linting tools to ensure the quality of the code. The easiest way to run these checks is to use grunt:

```
npm install
npm test
```

The karma task will try to open Chrome as a browser in which to run the tests. Make sure this is available or change the configuration in test\test.config.js

#Usage

## Bower

This project does not directly support bower. If you are using wiredep, you can dd the following to your 
bower.json file to allow wiredep to use this directive.

```json
  "overrides": {
    "angular-date-time-input": {
      "main": [
        "src/dateTimeInput.js",
      ],
      "dependencies": {
        "angular": "^1.x",
        "moment": "^2.x"
      }
    }
  }
```

## NPM
We use npm for dependency management. Add the following to your package

```shell
npm install angular-date-time-input --save
```
This will copy the angular-date-time-input files into your node_modules folder, along with its dependencies.

Load the script files in your application:
```html
<script type="text/javascript" src="node_modules/moment/moment.js"></script>
<script type="text/javascript" src="node_modules/angular/angular.js"></script>
<script type="text/javascript" src="node_modules/angular-date-time-input/src/js/dateTimeInput.js"></script>
```

Add this module as a dependency to your application module:

```html
var myAppModule = angular.module('MyApp', ['ui.dateTimeInput'])
```

Apply the directive to your form elements:

```html
<input data-date-time-input="YYYY-MMM-DD" />
```

## Options

The value of the date-time-input attribute is the format the date values will be displayed.

Nota bene: The value saved in the model is, by default, a JavaScript ```Date``` object, not a string.
This can result in differences between what is seen in the model and what is displayed.

### date-time-input

This option controls the way the date is displayed in the view, not the model.

```html
<input data-date-time-input="YYYY-MMM-DD" />
```
See MomentJS documentation for valid formats.

### date-formats

This option defines additional input formats that will be accepted. 

```html
<input ... data-date-formats="['YYYY-MMM-DD']" />
```

Nota bene: Parsing multiple formats is considerably slower than parsing a single format. 
If you can avoid it, it is much faster to parse a single format.

See [MomentJS documentation] (http://momentjs.com/docs/#/parsing/string-formats) for more information.

### date-parse-strict

This option enables/disables strict parsing of the input formats. 

```html
<input ... data-date-parse-strict="false" />
```

### model-type

```html
<input ... data-model-type="Date | moment | milliseconds | [custom format]" />
```

Default: ```'Date'```

Specifies the data type to use when storing the selected date in the model. 

Accepts any string value, but the following values have special meaning (these values are case sensitive) :
 * ```'Date'``` stores a Date instance in the model. Will accept Date, moment, milliseconds, and ISO 8601 strings as initial input from the model 
 * ```'moment'``` stores a moment instance in the model. Accepts the same initial values as ```Date```
 * ```'milliseconds'``` store the epoch milliseconds (since 1-1-1970) in the model. Accepts the same initial values as ```Date```

Any other value is considered a custom format string. 

##Contributing

See [Contributing] (contributing.md) document

## License

angular-date-time-input is released under the MIT license and is copyright 2016 Knight Rider Consulting, Inc.. Boiled down to smaller chunks, it can be described with the following conditions.

## It requires you to:

* Keep the license and copyright notice included in angular-date-time-input's CSS and JavaScript files when you use them in your works

## It permits you to:

* Freely download and use angular-date-time-input, in whole or in part, for personal, private, company internal, or commercial purposes
* Use angular-date-time-input in packages or distributions that you create
* Modify the source code
* Grant a sublicense to modify and distribute angular-date-time-input to third parties not included in the license

## It forbids you to:

* Hold the authors and license owners liable for damages as angular-date-time-input is provided without warranty
* Hold the creators or copyright holders of angular-date-time-input liable
* Redistribute any piece of angular-date-time-input without proper attribution
* Use any marks owned by Knight Rider Consulting, Inc. in any way that might state or imply that Knight Rider Consulting, Inc. endorses your distribution
* Use any marks owned by Knight Rider Consulting, Inc. in any way that might state or imply that you created the Knight Rider Consulting, Inc. software in question

## It does not require you to:

* Include the source of angular-date-time-input itself, or of any modifications you may have made to it, in any redistribution you may assemble that includes it
* Submit changes that you make to angular-date-time-input back to the angular-date-time-input project (though such feedback is encouraged)

The full angular-date-time-input license is located [in the project repository](https://github.com/dalelotts/angular-date-time-input/blob/master/LICENSE) for more information.


## Donating
Support this project and other work by Dale Lotts via [gittip][gittip-dalelotts].

[![Support via Gittip][gittip-badge]][gittip-dalelotts]

[gittip-badge]: https://rawgithub.com/twolfson/gittip-badge/master/dist/gittip.png
[gittip-dalelotts]: https://www.gittip.com/dalelotts/

[license-image]: http://img.shields.io/badge/license-MIT-blue.svg?style=flat
[license-url]: LICENSE

