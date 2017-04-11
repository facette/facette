app.controller('ShowGraphController', function($rootScope, $routeParams, $scope, browseCollection, timeRange) {
    $scope.graph = {
        id: $routeParams.id,
        options: angular.extend({frame: true}, browseCollection.getGlobalOptions(null))
    };

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
