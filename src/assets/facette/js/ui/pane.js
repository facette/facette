angular.module('facette.ui.pane', [])

.directive('ngPane', function() {
    return {
        restrict: 'A',
        scope: {},
        controller: 'PaneController'
    };
})

.controller('PaneController', function($scope, $rootScope, $element) {
    $rootScope.steps = {};

    // Define scope functions
    $rootScope.switch = function(step) {
        var element = angular.element('[ng-pane-step="' + step + '"]').show();
        element.siblings('[ng-pane-step]').hide();

        $scope.$parent.step = step;
    };

    // Add class on pane element
    $element.addClass('pane');

    // Switch to first pane by default
    $scope.$evalAsync(function() { $rootScope.switch(1); });
})

.directive('ngPaneStep', function($rootScope) {
    return {
        require: '^ngPane',
        restrict: 'A',
        scope: {},
        link: function(scope, element, attrs) {
            $rootScope.steps[attrs.ngPaneStep] = element;
        }
    };
});
