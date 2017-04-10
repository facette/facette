angular.module('facette.ui.notify', [])

.directive('notify', function() {
    return {
        restrict: 'E',
        replace: true,
        scope: {},
        controller: 'NotifyController',
        templateUrl: 'templates/notify.html'
    };
})

.controller('NotifyController', function($rootScope, $scope, $timeout) {
    $scope.active = false;
    $scope.message = null;
    $scope.type = null;
    $scope.icon = null;

    $scope.reset = function() {
        $scope.active = false;

        $timeout(function() {
            $scope.message = null;
            $scope.type = null;
            $scope.icon = null;
        }, 250);
    };

    // Attach events
    $rootScope.$on('Notify', function(e, message, options) {
        options = options || {};

        $scope.active = true;
        $scope.message = message;
        $scope.type = options.type || null;
        $scope.icon = options.icon || null;

        $timeout(function() {
            $scope.reset();
        }, 5000);
    });
});
