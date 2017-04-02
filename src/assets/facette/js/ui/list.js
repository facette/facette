angular.module('facette.ui.list', [])

.directive('list', function() {
    return {
        restrict: 'E',
        replace: true,
        transclude: true,
        scope: {},
        templateUrl: 'templates/list.html'
    };
})

.directive('listrow', function() {
    return {
        require: '^list',
        restrict: 'E',
        replace: true,
        transclude: true,
        scope: {},
        templateUrl: 'templates/listrow.html'
    };
})

.directive('listcolumn', function() {
    return {
        require: '^listrow',
        restrict: 'E',
        replace: true,
        transclude: true,
        scope: {},
        templateUrl: 'templates/listcolumn.html'
    };
});
