app.controller('AdminListController', function($q, $rootScope, $routeParams, $scope, $timeout, $translate, library,
    catalog, providers) {

    var factory;

    $scope.section = $routeParams.section;
    $scope.state = stateLoading;
    $scope.items = [];
    $scope.templates = ($scope.section == 'collections' || $scope.section == 'graphs') &&
        $routeParams.templates !== undefined;

    $scope.form = {
        search: ''
    };

    $scope.page = 1;
    $scope.limit = pagingLimit;

    if (catalogSections.indexOf($scope.section) != -1) {
        factory = catalog;
    } else if (librarySections.indexOf($scope.section) != -1) {
        factory = library;
    } else if ($scope.section == 'providers') {
        factory = providers;
    } else {
        return;
    }

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
            query.filter = 'glob:*' + $scope.form.search + '*';
        }

        if (factory !== catalog) {
            query.fields = 'id,name,description,created,modified';

            if ($scope.section == 'collections' || $scope.section == 'graphs') {
                query.kind = $routeParams.templates ? 'template' : 'raw';
                query.fields += ',link';
            } else if ($scope.section == 'providers') {
                query.fields += ',enabled';
            }
        }

        factory.list(query, function(data, headers) {
            $scope.items = data;
            $scope.total = parseInt(headers('X-Total-Records'), 10);
            $scope.state = stateOK;
        }, function() {
            $scope.state = stateError;
        });
    };

    $scope.reset = function() {
        $scope.form.search = '';
    };

    $scope.clone = function(item) {
        $rootScope.showModal({
            type: dialogTypePrompt,
            message: 'label.items_name',
            value: item.name + ' (clone)',
            labels: {
                validate: 'label.items_clone'
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

    $scope.remove = function(item) {
        $rootScope.showModal({
            type: dialogTypeConfirm,
            message: 'mesg.items_remove',
            args: {
                name: item.name
            },
            labels: {
                validate: 'label.items_remove'
            },
            danger: true
        }, function(data) {
            if (data === undefined) {
                return;
            }

            factory.delete({
                type: $scope.section,
                id: item.id
            }, function() {
                $scope.refresh();
            });
        });
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

    $scope.formatBasicTooltip = function(entry) {
        var defer = $q.defer();

        $translate(['label.identifier']).then(function(data) {
            defer.resolve('<span>' + data['label.identifier'] + '</span> ' + entry.id);
        });

        return defer.promise;
    };

    $scope.formatCatalogTooltip = function(name) {
        var defer = $q.defer();

        $q.all([
            $translate(['label.providers']),
            catalog.get({type: $scope.section, name: name}).$promise
        ]).then(function(data) {
            defer.resolve('<span>' + data[0]['label.providers'] + '</span> ' + data[1].providers.join(', '));
        });

        return defer.promise;
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
            $scope.refresh();
        }, 500);
    });

    // Load items
    $scope.refresh();
});
