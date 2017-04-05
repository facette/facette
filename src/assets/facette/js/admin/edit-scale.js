app.controller('AdminEditScaleController', function($routeParams, $scope, adminEdit) {
    $scope.section = 'scales';
    $scope.id = $routeParams.id;

    // Define scope functions
    $scope.cancel = function(force) { adminEdit.cancel($scope, force); };
    $scope.reset = function() { adminEdit.reset($scope); };

    $scope.save = function() {
        adminEdit.save(
            $scope,
            function(data) { data.value = parseFloat(data.value); },
            function(item) { return Boolean(item.value); }
        );
    };

    $scope.remove = function(list, entry) { adminEdit.remove($scope, list, entry); };

    // Register watchers
    adminEdit.watch($scope);

    // Initialize scope
    adminEdit.load($scope, function() {
        // Select first field
        $scope.$applyAsync(function() { angular.element('.pane :input:visible:first').select(); });
    });
});
