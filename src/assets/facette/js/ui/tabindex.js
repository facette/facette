angular.module('facette.ui.tabindex', [])

.directive('tabindex', function() {
    return {
        restrict: 'A',
        link: function(scope, element) {
            if (element.closest('label').length === 0) {
                return;
            }

            element.on('keypress', function(e) {
                if (e.which == 13 || e.which == 32) {
                    angular.element('#' + element.attr('for')).trigger('click');
                    e.preventDefault();
                }
            });
        }
    };
});
