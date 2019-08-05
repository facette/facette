app.controller('AdminEditCollectionController', function($q, $routeParams, $scope, $timeout, $translate, adminEdit,
    bulk, globalHotkeys, library, libraryAction) {

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
                endpoint: 'library/graphs/' + entry.graph,
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
                    keys = keys.concat(Array.from(input.matchAll(templateRegexp), function(m) { return m[1]; }));
                }
            });

            // Parse graphs for attribute fields
            angular.forEach($scope.item.entries, function(entry) {
                angular.forEach(entry.attributes, function(attr) {
                    keys = keys.concat(Array.from(attr.matchAll(templateRegexp), function(m) { return m[1]; }));
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
    $scope.cancel = function(force) { adminEdit.cancel($scope, force); };
    $scope.delete = function() { adminEdit.delete($scope, {id: $scope.id, name: $scope.itemRef.name}); };
    $scope.reset = function() { adminEdit.reset($scope); fetchGraphs(); };

    $scope.save = function(go) {
        adminEdit.save($scope, function(data) {
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
        }, go);
    };

    $scope.remove = function(list, entry) {
        adminEdit.remove($scope, list, entry);

        // Update template status
        updateTemplate();
    };

    $scope.selectGraph = function(e, data) {
        angular.extend($scope.graph, data);
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
            graph: $scope.graph.id,
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
        $scope.$applyAsync(function() { angular.element('#graph input').val('').focus(); });
    };

    $scope.editGraph = function(entry) {
        var idx = $scope.item.entries.indexOf(entry);
        if (idx == -1) {
            return;
        }

        $scope.graph = angular.extend({index: idx, id: entry.graph}, $scope.graphData[entry.graph]);

        $scope.$applyAsync(function() {
            angular.element('#graph input').val($scope.graphData[entry.graph].name).select();
        });
    };

    $scope.toggleGraph = function(entry) {
        if (!entry.options) {
            entry.options = {enabled: false};
        } else {
            entry.options.enabled = !entry.options.enabled;
        }
    };

    $scope.editGraphEntry = function(entry) {
        if (entry === null) {
            $scope.graphEntryEdit = null;
            $scope.entryOptions = null;
            $scope.entryAttrs = null;
            return;
        }

        var entryAttrs = angular.copy(entry.attributes || {});
        angular.forEach($scope.graphData[entry.graph].templateKeys, function(key) {
            if (!entryAttrs[key]) {
                entryAttrs[key] = '';
            }
        });

        $scope.entryOptions = angular.copy(entry.options || {});
        $scope.entryAttrs = entryAttrs;
        $scope.graphEntryEdit = entry;
    };

    $scope.setGraphEntry = function() {
        var idx = $scope.item.entries.indexOf($scope.graphEntryEdit);
        if (idx == -1) {
            return;
        }

        var entry = $scope.item.entries[idx];
        entry.options = angular.extend(entry.options || {}, $scope.entryOptions);
        entry.attributes = angular.extend(entry.attributes || {}, $scope.entryAttrs);

        $scope.editGraphEntry(null);
    };

    $scope.editAttrs = function(entry) {
        if (entry === null) {
            $scope.attrsEdit = null;
            $scope.entryAttrs = null;
            return;
        }

        var entryAttrs = angular.copy($scope.item.attributes || {});
        angular.forEach($scope.graphData, function(entry) {
            angular.forEach(entry.templateKeys, function(key) {
                if (!entryAttrs[key]) {
                    entryAttrs[key] = '';
                }
            });
        });

        $scope.entryAttrs = entryAttrs;
        $scope.attrsEdit = entry;

        // Select first field
        $timeout(function() { angular.element('.pane .keylist .value:first :input').select(); }, 0);
    };

    $scope.setAttrs = function() {
        $scope.item.attributes = angular.copy($scope.entryAttrs);
        $scope.editAttrs(null);
    };

    $scope.selectParent = function(e, data) {
        $scope.item.parent = data.id;
    };

    $scope.removeParent = function() {
        $scope.item.parent = null;
        $scope.$applyAsync(function() { angular.element('#parent input').val(''); });
    };

    $scope.selectTemplate = function(e, data) {
        $scope.item.link = data;
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
    adminEdit.watch($scope, function(newValue, oldValue) {
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
                        $scope.$applyAsync(function() { angular.element('#template input').val(data.name); });
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
    adminEdit.load($scope, function() {
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
                var defer = $q.defer();

                library.list({
                    type: 'graphs',
                    fields: 'id,name,options,template',
                    filter: 'glob:*' + term + '*'
                }, function(data, headers) {
                    defer.resolve({
                        entries: data.map(function(a) { return {label: a.name, value: a}; }),
                        total: parseInt(headers('X-Total-Records'), 10)
                    });
                }, function() {
                    defer.reject();
                });

                return defer.promise;
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
                var defer = $q.defer();

                library.list({
                    type: 'collections',
                    kind: 'template',
                    fields: 'id,name',
                    filter: 'glob:*' + term + '*'
                }, function(data, headers) {
                    defer.resolve({
                        entries: data.map(function(a) { return {label: a.name, value: a.id}; }),
                        total: parseInt(headers('X-Total-Records'), 10)
                    });
                }, function() {
                    defer.reject();
                });

                return defer.promise;
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
                var entries = [];

                angular.forEach(collections, function(entry) {
                    entries.push({label: entry.name, value: entry});
                });

                defer.resolve({
                    entries: entries,
                    total: entries.length
                });
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
                $scope.$applyAsync(function() { angular.element('#parent input').val(data.name); });
            });
        }
    });

    // Register global hotkeys
    globalHotkeys.register($scope);
});
