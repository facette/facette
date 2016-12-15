app.controller('BrowseSearchController', function($location, $scope, $window, BrowseCollection, libraryAction) {

    $scope.collections = {};
    $scope.collectionsLoaded = false;

    // Register scope functions
    $scope.searchHandler = function(term) {
        return libraryAction.search({
            types: ['collections', 'graphs'],
            terms: {
                name: 'glob:*' + term + '*',
                template: false
            },
            limit: pagingLimit
        }).$promise;
    };

    $scope.searchSelect = function(data) {
        $location.path('browse/' + data.originalObject.type + '/' + data.originalObject.value.id);
    };

    // Handle tree state save
    $scope.$on('$locationChangeStart', BrowseCollection.saveTreeState);
    angular.element($window).on('beforeunload', BrowseCollection.saveTreeState);

    // Load collections tree
    BrowseCollection.injectTree($scope);
});
