angular.module('facette.ui.search', [])

.directive('search', function() {
    return {
        restrict: 'E',
        replace: true,
        scope: {
            name: '@',
            icon: '@',
            placeholder: '@',
            ngModel: '='
        },
        controller: 'SearchController',
        templateUrl: 'templates/search.html'
    };
})

.controller('SearchController', function($scope) {
    $scope.hasFocus = false;

    $scope.handleFocus = function(e) {
        $scope.hasFocus = e.type == 'focus';
    };

    $scope.handleKey = function(e) {
        if (e.which != 27) { // <Escape>
            return;
        }

        if (!e.target.value) {
            $scope.$evalAsync(function() { e.target.blur(); });
        } else {
            $scope.ngModel = '';
        }

        e.preventDefault();
    };
});
