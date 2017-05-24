app.factory('browseCollection', function($location, $routeParams, $timeout, library, storage) {
    return {
        getGlobalOptions: function(scope) {
            var options = {};

            angular.forEach($routeParams, function(value, key) {
                switch (key) {
                case 'start':
                case 'end':
                case 'range':
                case 'time':
                    if (key == 'start' || key == 'end') {
                        key += '_time';
                    }

                    options[key] = value;
                    if (scope) {
                        scope[key] = options[key];
                    }
                    break;

                case 'refresh':
                    options.refresh_interval = parseInt(value, 10);
                    if (scope) {
                        scope.refreshInterval = options.refresh_interval;
                    }
                    break;
                }
            });

            return options;
        },

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

            var data = storage.get('browse-collection_tree', 'data');
            if (data) {
                scope.collections = data;
                scope.collectionsLoaded = true;
            }

            library.collectionTree({parent: scope.id || null}, function(data) {
                scope.collections = data;
                scope.collectionsLoaded = true;

                storage.set('browse-collection_tree', 'data', data);

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
        },

        watchGraphOptions: function(scope, key) {
            scope.$watch(key, function(newValue, oldValue) {
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
        }
    };
});
