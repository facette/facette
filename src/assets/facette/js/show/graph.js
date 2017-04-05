app.controller('ShowGraphController', function($rootScope, $routeParams, $scope, browseCollection) {
    $scope.graph = {
        id: $routeParams.id,
        options: angular.extend({frame: true}, browseCollection.getGlobalOptions(null))
    };

    // Set root scope loaded (no '$includeContentLoaded' event triggered on 'show' route)
    $rootScope.loaded = true;
});
