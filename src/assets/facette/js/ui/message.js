angular.module('facette.ui.message', [])

.directive('message', function() {
    return {
        restrict: 'E',
        replace: true,
        transclude: true,
        scope: {
            icon: '@',
            type: '@'
        },
        templateUrl: 'templates/message.html'
    };
});
