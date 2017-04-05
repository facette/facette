app.controller('AdminEditCollectionController', function($q, $routeParams, $scope, $timeout, $translate, AdminEdit,
    bulk, library, libraryAction) {

    $scope.section = 'collections';
    $scope.id = $routeParams.id;
    $scope.linked = $scope.id == 'link';
    $scope.tab = 0;

    $scope.graphFetched = false;
    $scope.graphData = {};
    $scope.hasTemplate = false;

    // Define helper functions
    function fetchGraphs() {
        var graphQuery = [],
            attrsQuery = [];

        $scope.graphFetched = false;

        angular.forEach($scope.item.entries, function(entry) {
            graphQuery.push({
                endpoint: 'library/graphs/' + entry.id,
                method: 'GET',
                params: {fields: 'id,name,options,template'}
            });
        });

        bulk.exec(graphQuery, function(data) {
            var graphData = {};

            angular.forEach(data, function(entry) {
                if (entry.status == 200) {
                    graphData[entry.data.id] = entry.data;

                    if (entry.data.template) {
                        attrsQuery.push({
                            endpoint: 'library/parse',
                            method: 'POST',
                            data: {id: entry.data.id, type: 'graphs'}
                        });
                    }

                    delete entry.data.id;
                }
            });

            $scope.graphFetched = true;

            if (attrsQuery.length > 0) {
                bulk.exec(attrsQuery, function(data) {
                    angular.forEach(data, function(entry, idx) {
                        graphData[attrsQuery[idx].data.id].templateKeys = entry.data;
                    });
                });
            }

            $scope.graphData = graphData;
        });
    }

    function updateTemplate() {
        if ($scope.linked) {
            libraryAction.parse({id: $scope.item.link, type: 'collections'}, function(data) {
                if (!$scope.item.attributes) {
                    $scope.item.attributes = {};
                }

                $scope.templateKeys = data;
            });
        } else {
            var keys = [],
                entries = [];

            // Parse templatable fields for attribute names
            entries.push($scope.item.description);
            if ($scope.item.options) {
                entries.push($scope.item.options.title);
            }

            angular.forEach(entries, function(input) {
                if (input) {
                    keys = keys.concat(input.matchAll(templateRegexp));
                }
            });

            // Parse graphs for attribute fields
            angular.forEach($scope.item.entries, function(entry) {
                angular.forEach(entry.attributes, function(attr) {
                    keys = keys.concat(attr.matchAll(templateRegexp));
                });
            });

            // Prepare attributes object and keys list
            keys.sort();
            keys = jQuery.unique(keys);

            $scope.templateKeys = keys;
            $scope.item.template = keys.length > 0;

            if ($scope.item.template) {
                delete $scope.item.alias;
            }
        }
    }

    function setDefaults(input) {
        return angular.extend({
            options: {
                grid_size: 1
            }
        }, input);
    }

    // Define scope functions
    $scope.cancel = function(force) { AdminEdit.cancel($scope, force); };
    $scope.reset = function() { AdminEdit.reset($scope); fetchGraphs(); };

    $scope.save = function() {
        AdminEdit.save($scope, function(data) {
            // Remove empty data from attributes
            angular.forEach(data.entries, function(entry) {
                angular.forEach(entry.attributes, function(value, key) {
                    if (!value) {
                        delete entry.attributes[key];
                    }
                });

                if (entry.options) {
                    if (entry.options.constants) {
                        entry.options.constants = parseFloatList(entry.options.constants);
                    }

                    if (entry.options.percentiles) {
                        entry.options.percentiles = parseFloatList(entry.options.percentiles);
                    }
                }
            });

            angular.forEach(data.attributes, function(value, key) {
                if (!value) {
                    delete data.attributes[key];
                }
            });
        }, function(item) {
            if ($scope.linked && !item.link) {
                return false;
            }

            return true;
        });
    };

    $scope.remove = function(list, entry) {
        AdminEdit.remove($scope, list, entry);

        // Update template status
        updateTemplate();
    };

    $scope.selectGraph = function(data) {
        if (!data || !data.originalObject) {
            return;
        }

        angular.extend($scope.graph, data.originalObject);
    };

    $scope.setGraph = function() {
        if (!$scope.item.entries) {
            $scope.item.entries = [];
        }

        if (!$scope.graphData[$scope.graph.id]) {
            var id = $scope.graph.id;

            $scope.graphData[id] = angular.copy($scope.graph);

            libraryAction.parse({id: id, type: 'graphs'}, function(data) {
                $scope.graphData[id].templateKeys = data;
            });
        }

        var graph = {
            id: $scope.graph.id,
            name: $scope.graph.name
        };

        if ($scope.graph.index !== undefined) {
            angular.extend($scope.item.entries[$scope.graph.index], graph);
        } else {
            $scope.item.entries.push(graph);
        }

        // Update template status
        updateTemplate();

        $scope.resetGraph();
    };

    $scope.resetGraph = function() {
        $scope.graph = {};

        $scope.$broadcast('angucomplete-alt:clearInput', 'graph');
        $scope.$applyAsync(function() { angular.element('#graph_value').focus(); }, 0);
    };

    $scope.editGraph = function(entry) {
        var idx = $scope.item.entries.indexOf(entry);
        if (idx == -1) {
            return;
        }

        $scope.graph = angular.extend({index: idx}, $scope.graphData[entry.id]);

        $scope.$broadcast('angucomplete-alt:changeInput', 'graph', $scope.graphData[entry.id].name);

        $scope.$applyAsync(function() { angular.element('#graph_value').select(); });
    };

    $scope.toggleGraph = function(entry) {
        if (!entry.options) {
            entry.options = {enabled: false};
        } else {
            entry.options.enabled = !entry.options.enabled;
        }
    };

    $scope.editOptions = function(entry) {
        if (entry === null) {
            $scope.optionsEdit = null;
            $scope.entryOptions = null;
            return;
        }

        $scope.entryOptions = angular.copy(entry.options || {});
        $scope.optionsEdit = entry;
    };

    $scope.setOptions = function() {
        var idx = $scope.item.entries.indexOf($scope.optionsEdit);
        if (idx == -1) {
            return;
        }

        $scope.item.entries[idx].options = angular.extend($scope.item.entries[idx].options || {}, $scope.entryOptions);
        $scope.editOptions(null);
    };

    $scope.editAttrs = function(entry, main) {
        if (entry === null) {
            $scope.attrsEdit = null;
            $scope.entryAttrs = null;
            return;
        }

        if (main) {
            $scope.entryAttrs = angular.copy($scope.item.attributes || {});

            angular.forEach($scope.graphData, function(entry) {
                angular.forEach(entry.templateKeys, function(key) {
                    if (!$scope.entryAttrs[key]) {
                        $scope.entryAttrs[key] = '';
                    }
                });
            });
        } else {
            $scope.entryAttrs = angular.copy(entry.attributes || {});

            angular.forEach($scope.graphData[entry.id].templateKeys, function(key) {
                if (!$scope.entryAttrs[key]) {
                    $scope.entryAttrs[key] = '';
                }
            });
        }

        $scope.attrsEdit = entry;

        // Select first field
        $timeout(function() { angular.element('.pane .keylist .value:first :input').select(); }, 0);
    };

    $scope.setAttrs = function(main) {
        if (main) {
            $scope.item.attributes = angular.copy($scope.entryAttrs);
        } else {
            var idx = $scope.item.entries.indexOf($scope.attrsEdit);
            if (idx == -1) {
                return;
            }

            $scope.item.entries[idx].attributes = angular.extend($scope.item.entries[idx].attributes || {},
                $scope.entryAttrs);
        }

        $scope.editAttrs(null, main);
    };

    $scope.selectParent = function(data) {
        if (!data || !data.originalObject || !data.originalObject.id) {
            return;
        }

        $scope.item.parent = data.originalObject.id;
    };

    $scope.removeParent = function() {
        $scope.item.parent = null;
        $scope.$broadcast('angucomplete-alt:clearInput', 'parent');
    };

    $scope.selectTemplate = function(data) {
        if (!data || !data.originalObject || !data.originalObject.id) {
            return;
        }

        $scope.item.link = data.originalObject.id;
    };

    $scope.switchTab = function(idx) {
        $scope.tab = idx;

        if (idx == 1) {
            library.list({
                type: 'collections',
                kind: 'raw',
                link: $scope.id,
                fields: 'id,name'
            }, function(data) {
                $scope.instances = data;
            });
        }
    };

    // Register watchers
    AdminEdit.watch($scope, function(newValue, oldValue) {
        if ($scope.step == 2 && !$scope.linked) {
            updateTemplate();
        } else if ($scope.linked) {
            if (!oldValue || newValue.link !== oldValue.link) {
                library.get({
                    type: 'collections',
                    id: newValue.link,
                    fields: 'name'
                }, function(data) {
                    // Restore selected template name
                    if (!oldValue) {
                        $scope.$broadcast('angucomplete-alt:changeInput', 'template', data.name);
                    }

                    updateTemplate();
                });
            }
        }
    });

    $scope.$watch('graphData', function(newValue, oldValue) {
        if (newValue === oldValue) {
            return;
        }

        var hasTemplate = false;
        angular.forEach(newValue, function(entry) {
            if (entry.template) {
                hasTemplate = true;
            }
        });

        if (hasTemplate !== $scope.hasTemplate) {
            $scope.hasTemplate = hasTemplate;
        }
    }, true);

    // Initialize scope
    AdminEdit.load($scope, function() {
        if ($scope.item.link) {
            $scope.linked = true;
        }

        if (!$scope.linked) {
            $scope.item = setDefaults($scope.item);
        }
        $scope.itemRef = angular.copy($scope.item);

        if (!$scope.linked) {
            $scope.selectedOptions = {};

            $scope.$watch('selectedOptions', function(newValue, oldValue) {
                // Handle select value changes
                if (!angular.equals(newValue, oldValue)) {
                    angular.forEach(newValue, function(entry, key) {
                        $scope.item.options[key] = entry.value;
                    });
                }
            }, true);

            $scope.graphsList = function(term) {
                return library.list({
                    type: 'graphs',
                    fields: 'id,name,options,template',
                    filter: 'glob:*' + term + '*'
                }).$promise;
            };

            $translate(['label.page_grid_1', 'label.page_grid_2', 'label.page_grid_3']).then(function(data) {
                var gridSizes = [];
                for (var i = 1; i <= 3; i++) {
                    gridSizes.push({name: data['label.page_grid_' + i], value: i});
                }
                $scope.collectionsGridSizes = gridSizes;

                applyOptions($scope.collectionsGridSizes, 'grid_size');
            });

            $scope.resetGraph();

            // Restore or set main options
            var applyOptions = function(list, key) {
                angular.forEach(list, function(entry) {
                    if (entry.value === $scope.item.options[key]) {
                        $scope.selectedOptions[key] = entry;
                    }
                });

                if (!$scope.selectedOptions[key]) {
                    $scope.selectedOptions[key] = list[0];
                }
            };

            // Trigger initial graph information retrieval
            fetchGraphs();
        } else {
            $scope.templateSources = function(term) {
                return library.list({type: 'collections', kind: 'template', fields: 'id,name',
                    filter: 'glob:*' + term + '*'}).$promise;
            };

            // Select first field
            $scope.$applyAsync(function() { angular.element('.pane :input:visible:first').select(); });
        }

        $scope.collectionsList = function(term) {
            var defer = $q.defer();

            library.list({
                type: 'collections',
                kind: 'raw',
                fields: 'id,name,parent',
                filter: 'glob:*' + term + '*'
            }).$promise.then(function(data) {
                var collections = {},
                    assocs = {};

                angular.forEach(data, function(entry) {
                    // Set parent association
                    if (entry.parent) {
                        if (!assocs[entry.parent]) {
                            assocs[entry.parent] = [];
                        }
                        assocs[entry.parent].push(entry.id);
                    }

                    // Set result entry
                    collections[entry.id] = entry;
                });

                // Clean up list from current children collections
                var stack = [$scope.id],
                    cur = null;

                while (stack.length > 0) {
                    cur = stack.shift();
                    if (assocs[cur]) {
                        stack = stack.concat(assocs[cur]);
                    }
                    delete collections[cur];
                }

                // Return cleaned up list
                var result = [];

                angular.forEach(collections, function(entry) {
                    result.push(entry);
                });

                defer.resolve(result);
            });

            return defer.promise;
        };

        // Resolve parent name
        if ($scope.item.parent) {
            library.get({
                type: 'collections',
                id: $scope.item.parent,
                fields: 'name'
            }, function(data) {
                $scope.$broadcast('angucomplete-alt:changeInput', 'parent', data.name);
            });
        }
    });
});
