angular.module('facette.ui.pane', [])

.directive('ngPane', function($rootScope) {
    return {
        restrict: 'A',
        scope: {},
        controller: 'PaneController'
    };
})

.controller('PaneController', function($scope, $rootScope, $element) {
    $rootScope.step = null;
    $rootScope.steps = {};

    // Define scope functions
    $rootScope.switch = function(step) {
        var element = angular.element('[ng-step="' + step + '"]').show();
        element.siblings('[ng-step]').hide();

        $rootScope.step = $scope.$parent.step = step;
    };

    // Add class on pane element
    $element.addClass('pane');

    // Switch to first pane by default
    $scope.$evalAsync(function() { $rootScope.switch(1); });
})

.directive('ngStep', function($rootScope) {
    return {
        require: '^ngPane',
        restrict: 'A',
        scope: {},
        link: function(scope, element, attrs) {
            $rootScope.steps[attrs.ngStep] = element;
        }
    };
});
