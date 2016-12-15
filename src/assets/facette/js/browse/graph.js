app.controller('BrowseGraphController', function($rootScope, $routeParams, $scope, $timeout, $window, BrowseCollection,
    TimeRange, bulk, library, storage) {

    $scope.section = $routeParams.section;
    $scope.id = $routeParams.id;
    $scope.state = stateLoading;
    $scope.collections = {};
    $scope.collectionsLoaded = false;
    $scope.graphs = [];

    $scope.form = {
        filter: ''
    };

    // Set range values
    $scope.rangeValues = timeRanges;

    // Register scope functions
    $scope.refresh = function() {
        $rootScope.$emit('RefreshGraph');
    };

    $scope.setRange = function(range) {
        if (range != 'custom') {
            $rootScope.$emit('ApplyGraphOptions', {range: '-' + range});
            return;
        }

        $rootScope.$emit('PromptTimeRange', function(time, range) {
            $scope.time = time;
            $scope.range = range;

            $rootScope.$emit('ApplyGraphOptions', {time: time, range: range});
        });
    };

    $scope.setRefresh = function() {
        $rootScope.showModal({
            type: dialogTypePrompt,
            message: 'label.graphs_refresh',
            note: 'label.refresh_interval_unit',
            value: $scope.refreshInterval,
            inputType: 'number',
            labels: {
                validate: 'label.graphs_refresh_set'
            }
        }, function(data) {
            if (data === undefined) {
                return;
            }

            $scope.refreshInterval = parseInt(data.value, 10);
            $rootScope.$emit('ApplyGraphOptions', {refresh_interval: $scope.refreshInterval});
        });
    };

    $scope.print = function() {
        $scope.$applyAsync(window.print);
    };

    $scope.setGrid = function(size) {
        $scope.gridSize = size;
        storage.set('browse-grid', $scope.section + '-' + $scope.id, size);
        $timeout(function() { $rootScope.$emit('RedrawGraph'); }, 0);
    };

    $scope.toggleLegends = function(state) {
        $rootScope.$emit('ToggleGraphLegend', null, null, state);
        $scope.showLegends = state;
    };

    $scope.handleGraphFocus = function(e, index) {
        angular.element('#graph' + index).next().toggleClass('focus', e.type == 'mouseenter');
    };

    $scope.resetFilter = function() {
        $scope.form.filter = '';
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
                baseMargin = parseInt(trees.first().find('.treelabel').css('padding-left'), 10);

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

    $scope.$watch('form.filter', function(newValue, oldValue) {
        if (angular.equals(newValue, oldValue)) {
            return;
        }

        if ($scope.filterTimeout) {
            $timeout.cancel($scope.filterTimeout);
            $scope.filterTimeout = null;
        }

        $scope.filterTimeout = $timeout(function() {
            var pause = {},
                count = 0;

            angular.forEach($scope.graphs, function(entry) {
                entry.hidden = entry.title.toLowerCase().indexOf(newValue.toLowerCase()) == -1;
                if (!entry.hidden) {
                    count++;
                }

                if (pause[entry.id] === undefined) {
                    pause[entry.id] = entry.hidden;
                }
            });

            angular.forEach(pause, function(state, id) {
                $rootScope.$emit('PauseGraphDraw', id, state);
            });

            $scope.noMatch = count === 0;
        }, 500);
    });

    // Attach events
    $rootScope.$on('PromptTimeRange', function(e, callback) {
        TimeRange.prompt(callback);
    });

    $scope.$on('GraphLoaded', function(e, idx, id) {
        var key = $scope.section + '_' + $scope.id,
            data = storage.get('browse-graph_state', key, {}),
            state = data[idx + '_' + id] || {};

        if (state.legendActive) {
            $rootScope.$emit('ToggleGraphLegend', idx, id, state);
        }
    });

    $scope.$on('GraphChanged', function(e, idx, id, state) {
        var key = $scope.section + '_' + $scope.id,
            data = storage.get('browse-graph_state', key, {});

        data[idx + '_' + id] = state;
        storage.set('browse-graph_state', key, data);
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

    // Load collections and graphs data
    if ($scope.id) {
        var query = {
            type: $scope.section,
            id: $scope.id,
            fields: 'id,name,entries.id,entries.options,entries.attributes,options,attributes'
        };

        // Always expand collections when browsing
        if ($scope.section == 'collections') {
            query.expand = 1;
        }

        library.get(query, function(data) {
            var title = data.options && data.options.title ? data.options.title : data.name;

            // Set page title
            $rootScope.setTitle([title]);

            // Set grid size
            $scope.gridSize = storage.get('browse-grid', $scope.section + '-' + $scope.id, data.options.grid_size || 1);
            if ($scope.gridSize < 1 || $scope.gridSize > 3) {
                $scope.gridSize = 1;
            }

            // Get graphs state for folding restore
            var graphsState = storage.get('browse-graph_state', $scope.section + '_' + $scope.id, {});

            if ($scope.section == 'collections') {
                $scope.parentID = data.parent;

                // Fill entries with collection graphs
                var graphs = [],
                    graphsResolve = {};

                angular.forEach(data.entries, function(entry, idx) {
                    if (entry.options && entry.options.enabled === false) {
                        return;
                    }

                    var graph = {
                        index: idx,
                        id: entry.id,
                        options: entry.options || {},
                        hidden: false
                    };

                    var state = graphsState[idx + '_' + entry.id];
                    if (state && typeof state.folded == 'boolean') {
                        graph.options.folded = state.folded;
                    }

                    // Keep useful graph data
                    if (entry.options && entry.options.title) {
                        graph.title = entry.options.title;
                    } else {
                        if (!graphsResolve[entry.id]) {
                            graphsResolve[entry.id] = [];
                        }

                        graphsResolve[entry.id].push(graph);
                    }

                    graph.attributes = angular.extend({}, data.attributes, entry.attributes) || null;

                    // Set custom embed path for collection view
                    graph.options.embeddable_path = 'collections/' + $scope.id + '/' + idx;

                    graphs.push(graph);
                });

                // Retrieve graphs titles information
                var infoQuery = [];
                angular.forEach(graphsResolve, function(entries, id) {
                    infoQuery.push({
                        endpoint: 'library/graphs/' + id,
                        method: 'GET',
                        params: {fields: 'id,name,options.title'}
                    });
                });

                if (infoQuery.length > 0) {
                    bulk.exec(infoQuery, function(data) {
                        angular.forEach(data, function(entry) {
                            if (entry.status != 200) {
                                return;
                            }

                            var title = entry.data.options && entry.data.options.title ?
                                entry.data.options.title : entry.data.name;

                            angular.forEach(graphsResolve[entry.data.id], function(entry) {
                                entry.title = title;
                            });
                        });
                    });
                }

                $scope.graphs = graphs;
            } else {
                // Set entries to single graph display
                $scope.graphs = [{
                    index: 0,
                    id: $scope.id,
                    options: {
                        embeddable_path: 'graphs/' + $scope.id
                    },
                    hidden: false,
                    title: title
                }];
            }

            // Set default refresh interval
            $scope.refreshInterval = data.options && data.options.refresh_interval ?
                data.options.refresh_interval : null;

            $scope.state = stateOK;
        }, function() {
            $scope.state = stateError;
        });
    }

    // Load collections tree
    BrowseCollection.injectTree($scope);
});
