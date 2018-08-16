app.controller('AdminInfoController', function($rootScope, $scope, globalHotkeys, version) {
    // Set page title
    $rootScope.setTitle(['label.info', 'label.admin_panel']);

    // Get information from back-end
    $scope.info = {};
    version.get(null, function(data) {
        $scope.info = data;
    });

    // Register global hotkeys
    globalHotkeys.register($scope);
});
