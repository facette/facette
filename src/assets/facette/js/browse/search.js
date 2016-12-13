app.controller('BrowseSearchController', function($location, $scope, $timeout, $window, BrowseCollection, libraryAction,
    storage) {

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

    // Register watchers
    $scope.$watch('collections', function(newValue, oldValue) {
        if (angular.equals(newValue, oldValue)) {
            return;
        }

        // Rerieve existing tree state
        var state = storage.get('browse-collection_tree', 'state', {});

        $timeout(function() {
            var trees = angular.element('#collections-tree .tree'),
                baseMargin = parseInt(angular.element('#collections-tree .treelabel:first').css('padding-left'), 10);

            trees.each(function(index, item) {
                var tree = angular.element(item);

                tree.children('.treeitem').children('.treelabel').css({
                    paddingLeft: parseInt(tree.closest('.treeitem').children('.treelabel')
                        .css('padding-left'), 10) + baseMargin
                });
            });

            // Restore tree state
            if (state) {
                angular.element('#collections-tree .treelabel').each(function(index, item) {
                    var label = angular.element(item),
                        href = label.attr('href');

                    if (state[href.substr(href.lastIndexOf('/') + 1)] === true) {
                        label.children('.toggle').trigger('click');
                    }
                });
            }

            $scope.collectionsLoaded = true;
        }, 250);
    });

    // Handle tree state save
    function saveTreeState() {
        var state = {};
        angular.element('#collections-tree [collapsed]').each(function(index, item) {
            var href = angular.element(item).children('.treelabel').attr('href'),
                id = href.substr(href.lastIndexOf('/') + 1);

            state[id] = item.getAttribute('collapsed') == 'false';
        });

        storage.set('browse-collection_tree', 'state', state);
    }

    $scope.$on('$locationChangeStart', saveTreeState);
    angular.element($window).on('beforeunload', saveTreeState);

    // Load collections tree
    BrowseCollection.injectTree($scope);
});
