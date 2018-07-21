app.controller('ShowCollectionController', function($rootScope, $routeParams, $scope, browseCollection, library,
    timeRange) {

    $scope.id = $routeParams.id;
    $scope.index = $routeParams.index;

    library.get({
        type: 'collections',
        id: $scope.id,
        fields: 'entries.graph,entries.attributes,entries.options,attributes',
        expand: 1
    }, function(data) {
        var entry = data.entries[$scope.index],
            graph = {
                id: entry.graph,
                attributes: entry.attributes || {},
                options: entry.options || {}
            };

        angular.extend(graph.options, browseCollection.getGlobalOptions(null));
        graph.options.frame = true;

        if (data.attributes) {
            angular.extend(graph.attributes, data.attributes);
        }

        $scope.graph = graph;

        browseCollection.watchGraphOptions($scope, 'graph.options');

        // Attach events
        var unregisterPromptTimerange = $rootScope.$on('PromptTimeRange', function(e, callback, data) {
            timeRange.prompt(callback, data);
        });

        $scope.$on('$destroy', function() {
            unregisterPromptTimerange();
        });

        // Set root scope loaded (no '$includeContentLoaded' event triggered on 'show' route)
        $rootScope.loaded = true;
    });
});
