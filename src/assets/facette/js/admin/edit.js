app.factory('adminEdit', function($location, $rootScope, $timeout, $translate, adminHelpers) {
    function confirmDelete(scope, item, message, args) {
        args = args || {};
        args.name = item.name;

        $rootScope.showModal({
            type: dialogTypeConfirm,
            message: message,
            args: args,
            labels: {
                validate: 'label.' + scope.section + '_delete'
            },
            danger: true
        }, function(data) {
            if (data === undefined) {
                return;
            }

            adminHelpers.getFactory(scope).delete({
                type: scope.section,
                id: item.id
            }, function() {
                // Check if an item was being edited. If so, return to items list
                if (scope.id) {
                    var locSearch = {};
                    if (scope.item && scope.item.template) {
                        locSearch.templates = 1;
                    }

                    $location.path('admin/' + scope.section + '/').search(locSearch);
                    return;
                }

                scope.refresh();
            });
        });
    }

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

        delete: function(scope, item) {
            if (scope.templates) {
                adminHelpers.getFactory(scope).listPeek({
                    type: scope.section,
                    kind: 'raw',
                    link: item.id,
                    fields: 'id'
                }, function(data, headers) {
                    var count = parseInt(headers('X-Total-Records'), 10);

                    if (count > 0) {
                        confirmDelete(scope, item, 'mesg.templates_delete', {count: count});
                    } else {
                        confirmDelete(scope, item, 'mesg.items_delete');
                    }
                });
            } else {
                confirmDelete(scope, item, 'mesg.items_delete');
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

        save: function(scope, transform, validate, go) {
            go = typeof go == 'boolean' ? go : false;

            if (!scope.item.name || validate && !validate(scope.item)) {
                scope.validated = true;
                return;
            }

            var locSearch = {};
            if (scope.item.template) {
                locSearch.templates = 1;
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
            var factory = adminHelpers.getFactory(scope);

            (scope.id != 'add' && scope.id != 'link' ? factory.update : factory.append)(data, function(_, header) {
                if (scope.itemTimeout) {
                    $timeout.cancel(scope.itemTimeout);
                    scope.itemTimeout = null;
                }

                $rootScope.preventUnload(false);

                if (go) {
                    var id = scope.id;
                    if (scope.id == 'add' || scope.id == 'link') {
                        var location = header('Location');
                        id = location.substr(location.lastIndexOf('/') + 1);
                    }

                    $location.path('browse/' + scope.section + '/' + id).search(locSearch);
                } else {
                    $location.path('admin/' + scope.section + '/').search(locSearch);
                }

                $timeout(function() {
                    $translate(['mesg.saved']).then(function(data) {
                        $rootScope.$emit('Notify', data['mesg.saved'], {type: 'success', icon: 'check-circle'});
                    });
                }, 500);
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
                        adminHelpers.getFactory(scope).list({
                            type: scope.section,
                            filter: newValue.name
                        }, function(data) {
                            scope.conflict.name = data.length > 0;
                        });
                    }

                    if (scope.aliasable && newValue.alias && newValue.alias !== oldValue.alias &&
                        newValue.alias !== scope.itemRef.alias) {

                        adminHelpers.getFactory(scope).getPeek({
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

            adminHelpers.getFactory(scope).get({
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
