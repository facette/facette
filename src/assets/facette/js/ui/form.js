angular.module('facette.ui.form', [])

.directive('ngEntersubmit', function() {
    return {
        restrict: 'A',
        scope: {
            ngEntersubmit: '&'
        },
        controller: function($scope, $element) {
            $element.on('keypress', function(e) {
                if (e.which == 13) {
                    $scope.ngEntersubmit();
                    $scope.$apply();
                }
            });
        }
    };
});
