angular.module('facette.ui.color', ['colorpicker.module'])

.directive('colorpicker', function() {
    return {
        restrict: 'E',
        replace: true,
        scope: {
            position: '@',
            value: '=?'
        },
        controller: 'ColorController',
        templateUrl: 'templates/color.html'
    };
})

.controller('ColorController', function($scope, $element) {
    // Trigger color picker
    $scope.handleKey = function(e) {
        switch (e.which) {
        case 8: // <Backspace>
            if (e.type == 'keypress') {
                $element.find('button.close').trigger('click');
                $scope.value = null;
                e.preventDefault();
            }

            break;

        case 27: // <Escape>
            if (e.type == 'keydown' && $element.find('.colorpicker-visible').length > 0) {
                $element.find('button.close').trigger('click');
                e.stopPropagation();
            }

            break;

        case 13: // <Enter>
        case 32: // <Space>
            if (e.type == 'keypress') {
                $element.find('[colorpicker]').trigger('click');
            } else if (e.type == 'keydown' && e.which == 13) {
                if (e.target.tagName == 'INPUT') {
                    $element.find('button.close').trigger('click');
                    e.preventDefault();
                }

                e.stopPropagation();
            }

            break;
        }
    };
});
