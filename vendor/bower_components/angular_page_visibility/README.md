## Angular Page Visibility: a Page Visibility API interface for Angular
`angular-page-visibility` is a tiny lib which integrate [Page Visibility API](https://developer.mozilla.org/en-US/docs/Web/Guide/User_experience/Using_the_Page_Visibility_API) with Angular. 

It is exposed as a scope, which `$broadcast`-s `pageFocused` and `pageBlurred` when page is focused / blurred.
For old browsers [not supporting page visibility API](http://caniuse.com/#feat=pagevisibility), it ignores it silently.

## Usage
To use `angular-page-visibility`, just inject it, then listen to the events.

```javascript
angular.module('app')
       .controller('MyController', function($scope, $pageVisibility) {
         $pageVisibility.$on('pageFocused', function(){
           // page is focused
         });

         $pageVisibility.$on('pageBlurred', function(){
           // page is blurred
         });
       });
```

## Installation

1) include script: script can be included via `bower` or downloading directly

  - via bower: 
  `$ bower install angular-page-visibility`

  - download directly
  ```html
  <script src="https://raw.githubusercontent.com/mz026/angular_page_visibility/v0.0.3/dist/page_visibility.min.js" type="text/javascript"></script>
  ```

2) include the module:

```javascript
angular.app('myApp', [ 'angular-page-visibility' ])

```

## Testing
to test `angular-page-visibility`, `grunt`, `karma` are needed.

1. `$ npm install`
2. `$ bower install`
3. `$ npm test`

## Licence:
MIT
