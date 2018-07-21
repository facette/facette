angular.module('facette.ui.include', [])

.directive('ngIncludeReplace', function() {
    return {
        require: 'ngInclude',
        restrict: 'A',
        link: function(scope, element) {
            element.replaceWith(element.children());
        }
    };
});
