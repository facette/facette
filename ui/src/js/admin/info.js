app.controller('AdminInfoController', function($rootScope, $scope, globalHotkeys, info) {
    // Set page title
    $rootScope.setTitle(['label.info', 'label.admin_panel']);

    // Get information from back-end
    $scope.info = {};
    info.get(null, function(data) {
        $scope.info = data;
    });

    // Register global hotkeys
    globalHotkeys.register($scope);
});
