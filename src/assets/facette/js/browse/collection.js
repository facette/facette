app.factory('BrowseCollection', function($timeout, library, storage) {
    return {
        injectTree: function(scope) {
            function applyState(base) {
                base = base || angular.element('#collections-tree');

                // Restore tree state
                var state = storage.get('browse-collection_tree', 'state', {});

                if (state) {
                    base.find('.treelabel').each(function(index, item) {
                        var label = angular.element(item),
                            href = label.attr('href');

                        if (state[href.substr(href.lastIndexOf('/') + 1)] === true) {
                            $timeout(function() {
                                label.children('.toggle').trigger('click');
                                applyState(label.next('.tree'));
                            }, 0);
                        }
                    });
                }
            }

            library.collectionTree({parent: scope.id || null}, function(data) {
                scope.collections = data;
                scope.collectionsLoaded = true;

                scope.$applyAsync(function() { applyState(); });
            });
        },
        saveTreeState: function() {
            var state = storage.get('browse-collection_tree', 'state', {});
            angular.element('#collections-tree [collapsed]').each(function(index, item) {
                var href = angular.element(item).children('.treelabel').attr('href'),
                    id = href.substr(href.lastIndexOf('/') + 1);

                var collapsed = item.getAttribute('collapsed') == 'true';
                if (state[id] || !collapsed) {
                    state[id] = !collapsed;
                }
            });

            storage.set('browse-collection_tree', 'state', state);
        }
    };
});
