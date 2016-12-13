angular.module('facette.ui.dialog', [])

.directive('dialog', function() {
    return {
        restrict: 'E',
        replace: true,
        transclude: true,
        scope: {},
        controller: 'DialogController',
        templateUrl: 'templates/dialog.html'
    };
})

.controller('DialogController', function($scope, $element) {
    $scope.handleMouse = function(e) {
        if (angular.element(e.target).hasClass('dialog')) {
            $scope.$applyAsync(function() { $element.find('.actions :input.cancel').trigger('click'); });
        }
    };
});
