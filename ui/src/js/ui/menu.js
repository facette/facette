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
            selectable: '=?'
        },
        link: function(scope, element, attrs, controller, transcludeFn) {
            // Check if menu item has sub-content
            transcludeFn(function(clone) {
                scope.hasContent = clone.length > 0;
            });

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
                if (e.type == 'focus') {
                    element.siblings('.focus').removeClass('focus');

                    if (scope.hasContent) {
                        element.addClass('focus');
                    }
                } else if (e.type == 'blur') {
                    var parentElement = element.parent().closest('.has-content'),
                        nextElement = angular.element(e.originalEvent.relatedTarget);

                    if (parentElement.length === 0 && (nextElement.length === 0 ||
                        nextElement.closest('.has-content').length === 0)) {
                        element.removeClass('focus');
                    } else if (!nextElement.parents('.has-content').is(parentElement)) {
                        parentElement.removeClass('focus');
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
