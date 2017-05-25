angular.module('facette.ui.menu', [])

.directive('menu', function() {
    return {
        restrict: 'E',
        replace: true,
        transclude: true,
        scope: {},
        controller: 'MenuController',
        templateUrl: 'templates/menu.html'
    };
})

.controller('MenuController', function($scope) {
    $scope.items = [];

    // Append new item to menu
    this.append = function(item) {
        $scope.items.push(item);
    };

    // Reset siblings states and select new item
    this.select = function(item) {
        angular.forEach($scope.items, function(entry) {
            entry.active = false;
        });

        item.active = true;
    };
})

.directive('menuitem', function($location, $timeout) {
    return {
        require: '^menu',
        restrict: 'E',
        replace: true,
        transclude: true,
        scope: {
            href: '@',
            target: '@',
            icon: '@',
            badge: '@',
            label: '@',
            name: '@',
            info: '@',
            infoDirection: '@',
            type: '@',
            drop: '=?',
            selectable: '=?'
        },
        link: function(scope, element, attrs, controller) {
            // Set defaults
            if (scope.selectable === undefined) {
                scope.selectable = true;
            }

            if (scope.infoDirection === undefined) {
                scope.infoDirection = 'bottom';
            }

            // Remove empty subcontent element
            if (!element.find('.subcontent').html()) {
                element.find('.subcontent').remove();
            }

            // Set active if URL matches
            if (attrs.href) {
                scope.active = scope.selectable && attrs.href && $location.path() == attrs.href;
            }

            // Watch for active state changes
            scope.$watch('active', function(active) {
                if (active) {
                    controller.select(scope);
                }
            });

            // Handle item selection
            scope.select = function() {
                scope.active = true;
            };

            // Handle focus events
            scope.handleFocus = function(e) {
                if (e.type == 'focus' && element.find('.subcontent').length > 0) {
                    element.addClass('focus');
                } else if (e.type == 'blur') {
                    var nextElement = angular.element(e.originalEvent.relatedTarget);

                    if (element.hasClass('focus') && nextElement.closest('.drop').length === 0 ||
                            element.closest('.subcontent').length > 0 &&
                            nextElement.closest('.subcontent').length === 0) {
                        element.closest('.drop').removeClass('focus');
                    }
                }
            };

            // Handle key events
            scope.handleKey = function(e) {
                if (e.which == 13) { // <Enter>
                    $timeout(function() { element.trigger('click'); });
                }
            };

            // Append menu item into parent menu
            controller.append(scope);
        },
        templateUrl: 'templates/menuitem.html'
    };
});
