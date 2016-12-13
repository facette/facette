/* globals define, module, require, angular, moment */
/* jslint vars:true */

/**
 * @license angular-date-time-input
 * (c) 2013-2015 Knight Rider Consulting, Inc. http://www.knightrider.com
 * License: MIT
 *
 *    @author Dale "Ducky" Lotts
 *    @since  2013-Sep-23
 */

;(function (root, factory) {
  'use strict'
  /* istanbul ignore if */
  if (typeof module !== 'undefined' && module.exports) {
    var ng = typeof angular === 'undefined' ? require('angular') : angular
    var mt = typeof moment === 'undefined' ? require('moment') : moment
    factory(ng, mt)
    module.exports = 'ui.bootstrap.datetimepicker'
    /* istanbul ignore next */
  } else if (typeof define === 'function' && /* istanbul ignore next */ define.amd) {
    define(['angular', 'moment'], factory)
  } else {
    factory(root.angular, root.moment)
  }
}(this, function (angular, moment) {
  'use strict'
  angular.module('ui.dateTimeInput', [])
    .service('dateTimeParserFactory', DateTimeParserFactoryService)
    .directive('dateTimeInput', DateTimeInputDirective)

  DateTimeParserFactoryService.$inject = []

  function DateTimeParserFactoryService () {
    return function ParserFactory (modelType, inputFormats, dateParseStrict) {
      var result
      // Behaviors
      switch (modelType) {
        case 'Date':
          result = handleEmpty(dateParser)
          break
        case 'moment':
          result = handleEmpty(momentParser)
          break
        case 'milliseconds':
          result = handleEmpty(millisecondParser)
          break
        default: // It is assumed that the modelType is a formatting string.
          result = handleEmpty(stringParserFactory(modelType))
      }

      return result

      function handleEmpty (delegate) {
        return function (viewValue) {
          if (angular.isUndefined(viewValue) || viewValue === '' || viewValue === null) {
            return null
          } else {
            return delegate(viewValue)
          }
        }
      }

      function dateParser (viewValue) {
        return momentParser(viewValue).toDate()
      }

      function momentParser (viewValue) {
        return moment(viewValue, inputFormats, moment.locale(), dateParseStrict)
      }

      function millisecondParser (viewValue) {
        return moment.utc(viewValue, inputFormats, moment.locale(), dateParseStrict).valueOf()
      }

      function stringParserFactory (modelFormat) {
        return function stringParser (viewValue) {
          return momentParser(viewValue).format(modelFormat)
        }
      }
    }
  }

  DateTimeInputDirective.$inject = ['dateTimeParserFactory']

  function DateTimeInputDirective (dateTimeParserFactory) {
    return {
      require: 'ngModel',
      restrict: 'A',
      scope: {
        'dateFormats': '='
      },
      link: linkFunction
    }

    function linkFunction (scope, element, attrs, controller) {
      // validation
      if (angular.isDefined(scope.dateFormats) && !angular.isString(scope.dateFormats) && !angular.isArray(scope.dateFormats)) {
        throw new Error('date-formats must be a single string or an array of strings i.e. date-formats="[\'YYYY-MM-DD\']" ')
      }

      if (angular.isDefined(attrs.modelType) && (!angular.isString(attrs.modelType) || attrs.modelType.length === 0)) {
        throw new Error('model-type must be "Date", "moment", "milliseconds", or a moment format string')
      }

      // variables
      var displayFormat = attrs.dateTimeInput || moment.defaultFormat

      var dateParseStrict = (attrs.dateParseStrict === undefined || attrs.dateParseStrict === 'true')

      var modelType = (attrs.modelType || 'Date')

      var inputFormats = [attrs.dateTimeInput, modelType].concat(scope.dateFormats).concat([moment.ISO_8601]).filter(unique)
      var formatterFormats = [modelType].concat(inputFormats).filter(unique)

      // Behaviors
      controller.$parsers.unshift(dateTimeParserFactory(modelType, inputFormats, dateParseStrict))

      controller.$formatters.push(formatter)

      controller.$validators.dateTimeInput = validator

      element.bind('blur', applyFormatters)

      // Implementation

      function unique (value, index, self) {
        return ['Date', 'moment', 'milliseconds', undefined].indexOf(value) === -1 &&
          self.indexOf(value) === index
      }

      function validator (modelValue, viewValue) {
        if (angular.isUndefined(viewValue) || viewValue === '' || viewValue === null) {
          return true
        }
        return moment(viewValue, inputFormats, moment.locale(), dateParseStrict).isValid()
      }

      function formatter (modelValue) {
        if (angular.isUndefined(modelValue) || modelValue === '' || modelValue === null) {
          return null
        }

        if (angular.isDate(modelValue)) {
          return moment(modelValue).format(displayFormat)
        } else if (angular.isNumber(modelValue)) {
          return moment.utc(modelValue).format(displayFormat)
        }
        return moment(modelValue, formatterFormats, moment.locale(), dateParseStrict).format(displayFormat)
      }

      function applyFormatters () {
        controller.$viewValue = controller.$formatters.filter(keepAll).reverse().reduce(applyFormatter, controller.$modelValue)
        controller.$render()

        function keepAll () {
          return true
        }

        function applyFormatter (memo, formatter) {
          return formatter(memo)
        }
      }
    }
  }
}))

