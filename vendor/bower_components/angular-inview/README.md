# InView Directive for AngularJS [![CircleCI](https://circleci.com/gh/thenikso/angular-inview.svg?style=svg)](https://circleci.com/gh/thenikso/angular-inview)

Check if a DOM element is or not in the browser current visible viewport.

```html
<div in-view="ctrl.myDivIsVisible = $inview" ng-class="{ isInView: ctrl.myDivIsVisible }"></div>
```

**This is a directive for AngularJS 1, support for Angular 2 is not in the works yet (PRs are welcome!)**

> Version 2 of this directive uses a lightwight embedded reactive framework and
it is a complete rewrite of v1

## Installation

### With npm

```
npm install angular-inview
```

### With bower

```
bower install angular-inview
```

## Setup

In your document include this scripts:

```html
<script src="/node_modules/angular/angular.js"></script>
<script src="/node_modules/angular-inview/angular-inview.js"></script>
```

In your AngularJS app, you'll need to import the `angular-inview` module:

```javascript
angular.module('myModule', ['angular-inview']);
```

Or with a module loader setup like Webpack/Babel you can do:

```javascript
import angularInview from 'angular-inview';

angular.module('myModule', [angularInview.name]);
```

## Usage

This module will define two directives: `in-view` and `in-view-container`.

### InView

```html
<any in-view="{expression using $inview}" in-view-options="{object}"></any>
```

The `in-view` attribute must contain a valid [AngularJS expression](http://docs.angularjs.org/guide/expression)
to work. When the DOM element enter or exits the viewport, the expression will
be evaluated. To actually check if the element is in view, the following data is
available in the expression:

- `$inview` is a boolean value indicating if the DOM element is in view.
  If using this directive for infinite scrolling, you may want to use this like
  `<any in-view="$inview&&myLoadingFunction()"></any>`.

- `$inviewInfo` is an object containint extra info regarding the event

  ```
  {
    changed: <boolean>,
    event: <DOM event>,
    element: <DOM element>,
    elementRect: {
      top: <number>,
      left: <number>,
      bottom: <number>,
      right: <number>,
    },
    viewportRect: {
      top: <number>,
      left: <number>,
      bottom: <number>,
      right: <number>,
    },
    direction: { // if generateDirection option is true
      vertical: <number>,
      horizontal: <number>,
    },
    parts: { // if generateParts option is true
      top: <boolean>,
      left: <boolean>,
      bottom: <boolean>,
      right: <boolean>,
    },
  }
  ```

  - `changed` indicates if the inview value changed with this event
  - `event` the DOM event that triggered the inview check
  - `element` the DOM element subject of the inview check
  - `elementRect` a rectangle with the virtual (considering offset) position of
    the element used for the inview check
  - `viewportRect` a rectangle with the virtual (considering offset) viewport
    dimensions used for the inview check
  - `direction` an indication of how the element has moved from the last event
    relative to the viewport. Ie. if you scoll the page down by 100 pixels, the
    value of `direction.vertical` will be `-100`
  - `parts` an indication of which side of the element are fully visible. Ie. if
    `parts.top=false` and `parts.bottom=true` it means that the bottom part of
    the element is visible at the top of the viewport (but its top part is
    hidden behind the browser bar)

An additional attribute `in-view-options` can be specified with an object value
containing:

- `offset`: An expression returning an array of values to offset the element position.

  Offsets are expressed as arrays of 4 values `[top, right, bottom, left]`.
  Like CSS, you can also specify only 2 values `[top/bottom, left/right]`.

  Values can be either a string with a percentage or numbers (in pixel).
  Positive values are offsets outside the element rectangle and
  negative values are offsets to the inside.

  Example valid values for the offset are: `100`, `[200, 0]`,
  `[100, 0, 200, 50]`, `'20%'`, `['50%', 30]`

- `viewportOffset`: Like the element offset but appied to the viewport. You may
  want to use this to shrink the virtual viewport effectivelly checking if your
  element is visible (i.e.) in the bottom part of the screen `['-50%', 0, 0]`.

- `generateDirection`: Indicate if the `direction` information should
  be included in `$inviewInfo` (default false).

- `generateParts`: Indicate if the `parts` information should
  be included in `$inviewInfo` (default false).

- `throttle`: a number indicating a millisecond value of throttle which will
  limit the in-view event firing rate to happen every that many milliseconds

### InViewContainer

Use `in-view-container` when you have a scrollable container that contains `in-view`
elements. When an `in-view` element is inside such container, it will properly
trigger callbacks when the container scrolls as well as when the window scrolls.

```html
<div style="height: 150px; overflow-y: scroll; position: fixed;" in-view-container>
	<div style="height: 300px" in-view="{expression using $inview}"></div>
</div>
```

## Examples

The following triggers the `lineInView` when the line comes in view:

```html
<li ng-repeat="t in testLines" in-view="lineInView($index, $inview, $inviewpart)">This is test line #{{$index}}</li>
```

**See more examples in the [`examples` folder](./examples).**

## Migrate from v1

Version 1 of this directive can still be installed with
`npm install angular-inview@1.5.7`. If you already have v1 and want to
upgrade to v2 here are some tips:

- `throttle` option replaces `debounce`. You can just change the name. Notice that
  the functioning has changed as well, a debounce waits until there are no more
  events for the given amount of time before triggering; throttle instead stabilizes
  the event triggering only once every amount of time. In practival terms this
  should not affect negativelly your app.
- `offset` and `viewportOffset` replace the old offset options in a more structured
  and flexible way. `offsetTop: 100` becomes `offset: [100, 0, 0, 0]`.
- `$inviewInfo.event` replaces `$event` in the expression.
- `generateParts` in the options has now to be set to `true` to have
  `$inviewInfo.parts` available.

## Contribute

1. Fork this repo
2. Setup your new repo with `npm install` and `npm install angular`
3. Edit `angular-inview.js` and `angular-inview.spec.js` to add your feature
4. Run `npm test` to check that all is good
5. Create a [PR](https://github.com/thenikso/angular-inview/pulls)

If you want to become a contributor with push access open an issue asking that
or contact the author directly.
