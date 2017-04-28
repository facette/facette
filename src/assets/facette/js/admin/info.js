app.controller('AdminInfoController', function($rootScope, $scope, info) {
    // Set page title
    $rootScope.setTitle(['label.info', 'label.admin_panel']);

    // Get information from back-end
    $scope.info = {};
    info.get(null, function(data) {
        $scope.info = data;
    });
});
