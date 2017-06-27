app.controller('AdminListController', function($location, $q, $rootScope, $route, $routeParams, $scope, $timeout,
    $translate, adminEdit, adminHelpers, catalog, globalHotkeys, hotkeys, providersAction) {

    $scope.section = $route.current.$$route._type;
    $scope.state = stateLoading;
    $scope.refreshing = false;
    $scope.items = [];
    $scope.templates = ($scope.section == 'collections' || $scope.section == 'graphs') &&
        $routeParams.templates !== undefined;

    $scope.form = {
        search: $routeParams.search || ''
    };

    $scope.page = 1;
    $scope.limit = pagingLimit;

    $scope.providersData = {};

    var factory = adminHelpers.getFactory($scope);

    // Set page title
    $rootScope.setTitle(['label.' + $scope.section, 'label.admin_panel']);

    // Define scope functions
    $scope.refresh = function(page) {
        var query;

        if (page !== undefined) {
            $scope.page = page;
        }

        query = {
            type: $scope.section,
            offset: ($scope.page - 1) * $scope.limit,
            limit: $scope.limit
        };

        if ($scope.form.search) {
            var parts = [];
            angular.forEach($scope.form.search.split(' '), function(part) {
                if (part.startsWith('origin:')) {
                    query.origin = part.substr(7);
                } else if (part.startsWith('source:')) {
                    query.source = part.substr(7);
                } else {
                    parts.push(part);
                }
            });

            query.filter = 'glob:*' + parts.join(' ') + '*';
        }

        if (factory !== catalog) {
            query.fields = 'id,name,description,created,modified';

            if ($scope.section == 'collections' || $scope.section == 'graphs') {
                query.kind = $routeParams.templates ? 'template' : 'raw';
                query.fields += ',link,alias';
            } else if ($scope.section == 'providers') {
                query.fields += ',enabled';
            }
        }

        if ($scope.state != stateLoading) {
            $scope.refreshing = true;
        }

        factory.list(query, function(data, headers) {
            $scope.items = data;
            $scope.total = parseInt(headers('X-Total-Records'), 10);
            $scope.state = stateOK;
            $timeout(function() { $scope.refreshing = false; }, 500);
        }, function() {
            $scope.state = stateError;
            $scope.refreshing = false;
        });
    };

    $scope.reset = function() {
        $scope.form.search = '';
    };

    $scope.clone = function(item) {
        $rootScope.showModal({
            type: dialogTypePrompt,
            message: 'label.' + $scope.section + '_name',
            value: item.name + '-clone',
            labels: {
                validate: 'label.' + $scope.section + '_clone'
            }
        }, function(data) {
            if (data === undefined) {
                return;
            }

            factory.append({
                inherit: item.id
            }, {
                type: $scope.section,
                name: data.value
            }, function() {
                $scope.refresh();
            });
        });
    };

    $scope.delete = function(item) {
        adminEdit.delete($scope, item);
    };

    $scope.toggle = function(entry) {
        factory.update({
            type: $scope.section,
            id: entry.id,
            enabled: !entry.enabled
        }, function() {
            entry.enabled = !entry.enabled;
        });
    };

    $scope.getProviders = function(name) {
        catalog.get({type: $scope.section, name: name}, function(data) {
            $scope.providersData[name] = data.providers;
        });
    };

    $scope.refreshProvider = function(entry) {
        providersAction.refresh({id: entry.id}, function() {
            entry.refreshing = true;
            $timeout(function() { entry.refreshing = false; }, 3000);
        });
    };

    // Register watchers
    $scope.$watch('form.search', function(newValue, oldValue) {
        if (angular.equals(newValue, oldValue)) {
            return;
        }

        if ($scope.searchTimeout) {
            $timeout.cancel($scope.searchTimeout);
            $scope.searchTimeout = null;
        }

        if (!newValue) {
            $scope.state = stateLoading;
        }

        // Trigger search apply
        $scope.searchTimeout = $timeout(function() {
            $location.skipReload().search('search', newValue || null).replace();
            $scope.refresh();
        }, 500);
    });

    // Load items
    $scope.refresh();

    // Register scope-specific and global hotkeys
    $translate('label.' + $scope.section + '_search').then(function(data) {
        hotkeys.bindTo($scope).add({
            combo: '/',
            description: data,
            callback: function() { angular.element('#search input').focus(); }
        });
    });

    globalHotkeys.register($scope);
});
