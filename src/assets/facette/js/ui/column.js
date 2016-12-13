angular.module('facette.ui.column', [])

.directive('columns', function() {
    return {
        restrict: 'E',
        replace: true,
        transclude: true,
        scope: {},
        template: '<div class="columns" ng-transclude=""></div>'
    };
})

.directive('column', function() {
    return {
        require: '^columns',
        restrict: 'E',
        replace: true,
        transclude: true,
        scope: {},
        template: '<div class="column" ng-transclude=""></div>'
    };
});
