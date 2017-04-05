app.controller('ShowCollectionController', function($rootScope, $routeParams, $scope, browseCollection, library) {
    $scope.id = $routeParams.id;
    $scope.index = $routeParams.index;

    library.get({
        type: 'collections',
        id: $scope.id,
        fields: 'entries.id,entries.attributes,options',
        expand: 1
    }, function(data) {
        var graph = data.entries[$scope.index] || {};
        graph.options = angular.extend(graph.options || {}, browseCollection.getGlobalOptions(null));
        graph.options.frame = true;

        $scope.graph = graph;

        // Set root scope loaded (no '$includeContentLoaded' event triggered on 'show' route)
        $rootScope.loaded = true;
    });
});
