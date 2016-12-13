/*globals angular, moment, $ */
(function () {
  'use strict';

  angular
    .module('demo.demoController', [])
    .controller('DemoController', demoController);

  demoController.$inject = ['$scope', '$log'];

  function demoController($scope, $log) {

    $scope.controllerName = 'demoController';

    $scope.data = {
      date1: new Date().getTime()
    };

    /* Bindable functions
    -----------------------------------------------*/
    $scope.setLocale = setLocale;

    moment.locale('en');

    function getLocale() {
      return moment.locale();
    }

    function setLocale(newLocale) {
      moment.locale(newLocale);
    }

  }

})();
