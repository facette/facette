app.controller('AdminEditUnitController', function($routeParams, $scope, AdminEdit) {
    $scope.section = 'units';
    $scope.id = $routeParams.id;

    // Define scope functions
    $scope.cancel = function(force) { AdminEdit.cancel($scope, force); };
    $scope.reset = function() { AdminEdit.reset($scope); };
    $scope.save = function() { AdminEdit.save($scope, null, function(item) { return Boolean(item.label); }); };
    $scope.remove = function(list, entry) { AdminEdit.remove($scope, list, entry); };

    // Register watchers
    AdminEdit.watch($scope);

    // Initialize scope
    AdminEdit.load($scope, function() {
        // Select first field
        $scope.$applyAsync(function() { angular.element('.pane :input:visible:first').select(); });
    });
});
