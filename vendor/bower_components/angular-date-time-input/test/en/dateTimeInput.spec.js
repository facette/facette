/* globals moment, module, describe, it, expect, beforeEach, inject */
/**
 * @license angular-date-time-input
 * (c) 2015 Knight Rider Consulting, Inc. http://www.knightrider.com
 * License: MIT
 *
 *    @author Dale "Ducky" Lotts
 *    @since  2013-Sep-23
 */

describe('date-time-input', function () {
  'use strict'
  var compiler
  var rootScope

  beforeEach(module('ui.dateTimeInput'))

  beforeEach(inject(function ($compile, $rootScope) {
    moment.locale('en')
    rootScope = $rootScope
    compiler = $compile
  }))

  describe('valid configuration', function () {
    it('requires ngModel', function () {
      var compile = function () {
        compiler('<input data-date-time-input="M/D/YYYY h:mm A"/>')(rootScope)
      }

      expect(compile).toThrow()
    })

    it('does NOT require a date format', function () {
      compiler('<input data-date-time-input data-ng-model="dateValue" />')(rootScope)
    })

    it('requires date-formats to be a string or array, not an expression', function () {
      var compile = function () {
        compiler('<input data-ng-model="dateValue" data-date-time-input="M/D/YYYY h:mm A" data-date-formats="YYYY-MM-DD"/>')(rootScope)
      }

      expect(compile).toThrow()
    })

    it('requires date-formats to be a string or array, not a number', function () {
      var compile = function () {
        compiler('<input data-ng-model="dateValue" data-date-time-input="M/D/YYYY h:mm A" data-date-formats="0"/>')(rootScope)
      }

      expect(compile).toThrow()
    })

    it('accepts valid format string value from model', function () {
      rootScope.dateValue = '2016-01-23T06:00:00.000Z'
      var element = compiler('<input data-date-time-input data-ng-model="dateValue" />')(rootScope)
      rootScope.$digest()
      expect(element.val()).toBe(moment('2016-01-23T06:00:00.000Z').format())
    })

    it('accepts a single string as date-format', function () {
      rootScope.dateValue = '2016-01-23T06:00:00.000Z'
      var element = compiler('<input data-date-time-input data-ng-model="dateValue" data-date-formats="\'YYYY-MM-DD\'"/>')(rootScope)
      rootScope.$digest()
      expect(element.val()).toBe(moment('2016-01-23T06:00:00.000Z').format())
    })

    it('accepts an array of strings as date-format', function () {
      rootScope.dateValue = '2016-01-23T06:00:00.000Z'
      var element = compiler('<input data-date-time-input data-ng-model="dateValue" data-date-formats="[\'YYYY-MM-DD\']"/>')(rootScope)
      rootScope.$digest()
      expect(element.val()).toBe(moment('2016-01-23T06:00:00.000Z').format())
    })
  })

  describe('has expected initial structure', function () {
    it('is a `<input>` element', function () {
      var element = compiler('<input data-date-time-input="M/D/YYYY h:mm A" data-ng-model="dateValue"/>')(rootScope)
      rootScope.$digest()
      expect(element.prop('tagName')).toBe('INPUT')
    })
  })

  describe('accepts', function () {
    it('input matching valid display format', function () {
      var element = compiler('<input data-date-time-input="D/M/YYYY" data-ng-model="dateValue"/>')(rootScope)
      rootScope.$digest()
      element.val('01/12/2016')
      element.trigger('input')
      element.trigger('blur')
      rootScope.$digest()
      expect(rootScope.dateValue).toEqual(moment('2016-12-01').toDate())
      expect(element.val()).toEqual('1/12/2016')
    })

    it('input matching valid date format (not matching display format)', function () {
      var element = compiler('<input data-date-time-input="D/M/YYYY" data-date-formats="\'YYYY-DD-MM\'" data-ng-model="dateValue"/>')(rootScope)
      rootScope.$digest()
      element.val('2016-01-12')
      element.trigger('input')
      element.trigger('blur')
      rootScope.$digest()
      expect(rootScope.dateValue).toEqual(moment('2016-12-01').toDate())
      expect(element.val()).toEqual('1/12/2016')
    })

    it('ISO 8601 formatted input that does not match any specified format', function () {
      var element = compiler('<input data-date-time-input="D/M/YYYY" data-date-formats="\'YYYY-DD-MM\'" data-ng-model="dateValue"/>')(rootScope)
      rootScope.$digest()
      element.val('2016-01-23T06:00:00.000Z')
      element.trigger('input')
      element.trigger('blur')
      rootScope.$digest()
      expect(rootScope.dateValue).toEqual(moment('2016-01-23T06:00:00.000Z').toDate())
      expect(element.val()).toEqual('23/1/2016')
    })

    it('"1/1/1070" and displays "12/31/1969 12:00 AM" (BUG?) if parsing is NOT strict"', function () {
      var element = compiler('<input data-date-time-input="M/D/YYYY h:mm A" data-date-parse-strict="false" data-ng-model="dateValue"/>')(rootScope)
      rootScope.$digest()
      element.val('1/1/1970')
      element.trigger('input')

      var expectedMoment = moment('1/1/1970', 'M/D/YYYY h:mm A', moment.locale(), false)
      expect(rootScope.dateValue).toEqual(expectedMoment.toDate())
      expect(element.val()).toEqual('1/1/1970')
      element.trigger('blur') // formatting happens on blur
      rootScope.$digest()
      expect(element.val()).toEqual(expectedMoment.format('M/D/YYYY h:mm A'))
    })
  })

  // ToDo: locale specific formats
  // ToDo: model format - Date, moment, string, unix

  describe('has existing ngModel value and display value', function () {
    it('is "1/1/1970 12:00 am if the ngModel value is set to 1/1/1970', function () {
      rootScope.dateValue = moment('1/1/1970 12:00 am', 'M/D/YYYY h:mm A').toDate()
      var element = compiler('<input data-date-time-input="M/D/YYYY h:mm A" data-ng-model="dateValue"/>')(rootScope)
      rootScope.$digest()
      expect(element.val()).toEqual('1/1/1970 12:00 AM')
    })
  })

  describe('rejects invalid input', function () {
    it('of "foo"', function () {
      var element = compiler('<input data-date-time-input="M/D/YYYY h:mm A" data-ng-model="dateValue"/>')(rootScope)
      rootScope.$digest()
      element.val('foo')
      element.trigger('input')
      expect(rootScope.dateValue).toBe(undefined)
      expect(element.val()).toEqual('foo')
      expect(element.hasClass('ng-invalid')).toBe(true)
      element.trigger('blur') // formatting happens on blur
      expect(element.val()).toEqual('')
      expect(element.hasClass('ng-invalid')).toBe(true)
    })
  })
})
