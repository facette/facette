angular.module('facette.ui.tab', [])

.directive('tabset', function() {
    return {
        restrict: 'E',
        replace: true,
        transclude: true,
        scope: {},
        controller: 'TabController',
        templateUrl: 'templates/tabset.html'
    };
})

.controller('TabController', function($scope) {
    $scope.tabs = [];

    // Append new tab to set
    this.append = function(tab) {
        $scope.tabs.push(tab);
    };

    // Reset siblings states and select new tab
    this.select = function(tab) {
        angular.forEach($scope.tabs, function(entry, idx) {
            entry.active = angular.equals(entry, tab);
        });
    };
})

.directive('tab', function($timeout) {
    return {
        require: '^tabset',
        restrict: 'E',
        replace: true,
        scope: {
            active: '=?',
            label: '@',
            href: '@'
        },
        link: function(scope, element, attrs, controller) {
            // Watch for active state changes
            scope.$watch('active', function(active) {
                if (active) {
                    controller.select(scope);
                }
            });

            // Handle tab selection
            scope.select = function() {
                scope.active = true;
            };

            // Handle keys events
            scope.handleKey = function(e) {
                if (e.which != 13) { // <Enter>
                    return;
                }

                e.stopPropagation();
                $timeout(function() { element.trigger('click'); }, 0);
            };

            // Append tab into parent set
            controller.append(scope);
        },
        templateUrl: 'templates/tab.html'
    };
});
