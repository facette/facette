angular.module('facette.ui.transition', [])

.directive('ngTransitionend', function($parse) {
    return {
        restrict: 'A',
        link: function(scope, element, attrs) {
            var callback = $parse(attrs.ngTransitionend);

            element.on('transitionend otransitionend webkitTransitionEnd', function(e) {
                scope.$apply(function() {
                    callback(scope, {$event: e});
                });
            });
        }
    };
});
