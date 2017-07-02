app.controller('BrowseSearchController', function($location, $q, $rootScope, $scope, $window, browseCollection,
    globalHotkeys, libraryAction) {

    $scope.collections = {};
    $scope.collectionsLoaded = false;

    // Set page title
    $rootScope.setTitle();

    // Register scope functions
    $scope.searchHandler = function(term) {
        var defer = $q.defer();

        libraryAction.search({
            types: ['collections', 'graphs'],
            terms: {
                name: 'glob:*' + term + '*',
                template: false
            },
            limit: pagingLimit
        }, function(data) {
            defer.resolve(data.map(function(a) {
                return {
                    label: a.name,
                    value: a,
                    note: a.type
                };
            }));
        }, function() {
            defer.reject();
        });

        return defer.promise;
    };

    $scope.searchSelect = function(e, data) {
        $location.path('browse/' + data.type + '/' + data.id);
    };

    // Handle tree state save
    $scope.$on('$locationChangeStart', browseCollection.saveTreeState);
    angular.element($window).on('beforeunload', browseCollection.saveTreeState);

    // Load collections tree
    browseCollection.injectTree($scope);

    // Register global hotkeys
    globalHotkeys.register($scope);
});
