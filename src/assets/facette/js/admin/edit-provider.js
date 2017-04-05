app.controller('AdminEditProviderController', function($routeParams, $scope, adminEdit, info) {
    $scope.section = 'providers';
    $scope.id = $routeParams.id;

    // Define scope functions
    $scope.cancel = function(force) { adminEdit.cancel($scope, force); };
    $scope.reset = function() { adminEdit.reset($scope); };
    $scope.save = function() { adminEdit.save($scope, null, function(item) { return Boolean(item.connector); }); };
    $scope.remove = function(list, entry) { adminEdit.remove($scope, list, entry); };

    $scope.addFilter = function() {
        if (!$scope.item.filters) {
            $scope.item.filters = [];
        }

        $scope.item.filters.push({
            action: $scope.filterActions[0],
            target: $scope.filterTargets[0]
        });
    };

    // Register watchers
    adminEdit.watch($scope);

    // Initialize scope
    adminEdit.load($scope, function() {
        // Select first field
        $scope.$applyAsync(function() { angular.element('.pane :input:visible:first').select(); });

        // Fill connectors list
        $scope.connectorTypes = [];
        info.get(null, function(data) {
            if (data.connectors) {
                $scope.connectorTypes = data.connectors;
            }
        });

        // Set filter actions and targets
        $scope.filterActions = filterActions;
        $scope.filterTargets = filterTargets;
    });
});
