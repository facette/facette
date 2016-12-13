app.controller('ShowGraphController', function($rootScope, $routeParams, $scope) {
    $scope.id = $routeParams.id;

    // Set root scope loaded (no '$includeContentLoaded' event triggered on 'show' route)
    $rootScope.loaded = true;
});
