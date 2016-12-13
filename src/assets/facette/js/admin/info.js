app.controller('AdminInfoController', function($rootScope, $scope, info) {
    // Set page title
    $rootScope.setTitle(['label.info', 'label.admin_panel']);

    // Get information from backend
    $scope.info = {};
    info.get(null, function(data) {
        $scope.info = data;
    });
});
