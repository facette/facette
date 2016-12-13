app.controller('ErrorController', function($rootScope, $scope) {
    // Set page title
    $rootScope.setTitle(['label.error']);

    // Get root scope loaded
    $rootScope.loaded = true;
});
