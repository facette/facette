app.factory('AdminEdit', function($location, $rootScope, $timeout, library, providers) {
    return {
        cancel: function(scope, force) {
            force = typeof force == 'boolean' ? force : false;

            if (force) {
                $rootScope.preventUnload(false);
            }

            var locSearch = {};
            if (scope.item && scope.item.template) {
                locSearch.templates = 1;
            }

            $location.path('admin/' + scope.section + '/').search(locSearch);
            if (force) {
                $location.replace();
            }
        },

        reset: function(scope, callback) {
            $rootScope.showModal({
                type: dialogTypeConfirm,
                message: 'mesg.items_reset',
                labels: {
                    validate: scope.section.endsWith('groups') ?
                        'label.groups_reset' : 'label.' + scope.section + '_reset'
                },
                danger: true
            }, function(data) {
                if (data === undefined) {
                    return;
                }

                scope.item = angular.copy(scope.itemRef);

                if (callback) {
                    callback();
                }
            });
        },

        save: function(scope, transform, validate) {
            if (!scope.item.name || validate && !validate(scope.item)) {
                scope.validated = true;
                return;
            }

            var locSearch = {};
            if (scope.item.template) {
                locSearch.templates = 1;
            }

            // Skip save is no change applied
            if (!scope.modified) {
                $location.path('admin/' + scope.section + '/').search(locSearch);
                return;
            }

            var data = angular.extend({type: scope.section}, scope.item);
            if (scope.id != 'add' && scope.id != 'link') {
                data.id = scope.id;
            }

            if (scope.cleanProperties) {
                scope.cleanProperties(data);
            }

            if (transform) {
                transform(data);
            }

            scope.conflict = {name: false, alias: name};
            scope.validated = true;

            // Prepare item data
            var factory = scope.section == 'providers' ? providers : library;

            (scope.id != 'add' && scope.id != 'link' ? factory.update : factory.append)(data, function() {
                if (scope.itemTimeout) {
                    $timeout.cancel(scope.itemTimeout);
                    scope.itemTimeout = null;
                }

                $rootScope.preventUnload(false);
                $location.path('admin/' + scope.section + '/').search(locSearch);
            });
        },

        remove: function(scope, list, entry) {
            var index = list.indexOf(entry);
            if (index == -1) {
                return;
            }

            list.splice(index, 1);
        },

        watch: function(scope, callback) {
            scope.aliasable = scope.section == 'collections' || scope.section == 'graphs';

            scope.$watch('item', function(newValue, oldValue) {
                if (scope.state != stateOK || angular.equals(newValue, oldValue)) {
                    return;
                }

                // Set modification flag
                var item = angular.copy(scope.item);
                if (scope.cleanProperties) {
                    scope.cleanProperties(item);
                }

                scope.modified = !angular.equals(item, scope.itemRef);

                if (scope.itemTimeout) {
                    $timeout.cancel(scope.itemTimeout);
                    scope.itemTimeout = null;
                }

                scope.itemTimeout = $timeout(function() {
                    $rootScope.preventUnload(scope.modified);

                    // Execute callback if any
                    if (callback) {
                        callback(newValue, oldValue);
                    }

                    // Reset conflict flag on name reset
                    if (!newValue.name || newValue.name === scope.itemRef.name) {
                        scope.conflict.name = false;
                    }

                    if (!newValue.alias || newValue.alias === scope.itemRef.alias) {
                        scope.conflict.alias = false;
                    }

                    if (!oldValue) {
                        return;
                    }

                    // Check for name and/or alias conflicts
                    if (newValue.name && newValue.name !== oldValue.name && newValue.name !== scope.itemRef.name) {
                        (scope.section == 'providers' ? providers : library).list({
                            type: scope.section,
                            filter: newValue.name
                        }, function(data) {
                            scope.conflict.name = data.length > 0;
                        });
                    }

                    if (scope.aliasable && newValue.alias && newValue.alias !== oldValue.alias &&
                        newValue.alias !== scope.itemRef.alias) {

                        library.getPeek({
                            type: scope.section,
                            id: newValue.alias
                        }, function(data) {
                            scope.conflict.alias = true;
                        }, function(data) {
                            scope.conflict.alias = false;
                        });
                    }
                }, 500);
            }, true);
        },

        load: function(scope, callback) {
            scope.conflict = {name: false, alias: false};
            scope.validated = false;

            // Set page title
            $rootScope.setTitle(['label.' + scope.section +
                (scope.id == 'add' || scope.id == 'link' ? '_new' : '_edit'), 'label.admin_panel']);

            // Set sorting control settings
            scope.listSortControl = {
                allowDuplicates: true,
                containment: 'tbody'
            };

            // Initialize new item
            if (scope.id == 'add' || scope.id == 'link') {
                scope.item = {};
                scope.itemRef = {};
                scope.state = stateOK;

                if (callback) {
                    callback();
                }

                return;
            }

            // Load existing item
            scope.state = stateLoading;

            (scope.section == 'providers' ? providers : library).get({
                type: scope.section,
                id: scope.id
            }, function(data) {
                data = data.toJSON();
                delete data.id;

                scope.item = angular.copy(data);
                scope.itemRef = data;
                scope.state = stateOK;

                if (callback) {
                    callback();
                }
            }, function(response) {
                scope.state = stateError;
                scope.notFound = response.status == 404;
            });
        }
    };
});
