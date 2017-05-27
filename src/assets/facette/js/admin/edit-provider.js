app.controller('AdminEditProviderController', function($routeParams, $scope, adminEdit, info) {
    $scope.section = 'providers';
    $scope.id = $routeParams.id;

    // Define scope functions
    $scope.cancel = function(force) { adminEdit.cancel($scope, force); };
    $scope.delete = function() { adminEdit.delete($scope, {id: $scope.id, name: $scope.itemRef.name}); };
    $scope.reset = function() { adminEdit.reset($scope); };
    $scope.save = function() { adminEdit.save($scope, null, function(item) { return Boolean(item.connector); }); };
    $scope.remove = function(list, entry) { adminEdit.remove($scope, list, entry); };

    $scope.addFilter = function() {
        var actionIdx = 0,
            targetIdx = 0;

        if (!$scope.item.filters) {
            $scope.item.filters = [];
        } else {
            var last = $scope.item.filters[$scope.item.filters.length-1];
            actionIdx = $scope.filterActions.indexOf(last.action);
            targetIdx = $scope.filterTargets.indexOf(last.target);
        }

        $scope.item.filters.push({action: $scope.filterActions[actionIdx], target: $scope.filterTargets[targetIdx]});
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
