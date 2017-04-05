app.controller('AdminEditUnitController', function($routeParams, $scope, adminEdit) {
    $scope.section = 'units';
    $scope.id = $routeParams.id;

    // Define scope functions
    $scope.cancel = function(force) { adminEdit.cancel($scope, force); };
    $scope.reset = function() { adminEdit.reset($scope); };
    $scope.save = function() { adminEdit.save($scope, null, function(item) { return Boolean(item.label); }); };
    $scope.remove = function(list, entry) { adminEdit.remove($scope, list, entry); };

    // Register watchers
    adminEdit.watch($scope);

    // Initialize scope
    adminEdit.load($scope, function() {
        // Select first field
        $scope.$applyAsync(function() { angular.element('.pane :input:visible:first').select(); });
    });
});
