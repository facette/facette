app.controller('ShowGraphController', function($location, $rootScope, $routeParams, $scope, browseCollection,
    timeRange) {

    $scope.graph = {
        id: $routeParams.id,
        options: browseCollection.getGlobalOptions(null)
    };

    $scope.$watch('graph.options', function(newValue, oldValue) {
        if (angular.equals(newValue, oldValue)) {
            return;
        }

        $location.skipReload()
            .search('start', newValue.start_time || null)
            .search('end', newValue.end_time || null)
            .search('time', newValue.time || null)
            .search('range', newValue.range || null)
            .replace();
    }, true);

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
