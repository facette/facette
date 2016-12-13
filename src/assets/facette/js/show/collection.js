app.controller('ShowCollectionController', function($rootScope, $routeParams, $scope, library) {
    $scope.id = $routeParams.id;
    $scope.index = $routeParams.index;

    // Set root scope loaded (no '$includeContentLoaded' event triggered on 'show' route)
    $rootScope.loaded = true;

    library.get({
        type: 'collections',
        id: $scope.id,
        fields: 'entries.id,entries.attributes',
        expand: 1
    }, function(data) {
        $scope.graph = data.entries[$scope.index] || {};
    });
});
