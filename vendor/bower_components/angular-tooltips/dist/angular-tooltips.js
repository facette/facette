(function() {
    'use strict';

    var directive = function ($timeout, $compile) {
        return {
            restrict: 'A',
            scope: {
                title: '@',
                fixedPosition: '=',
                titleClass: '@',
            },
            link: function ($scope, element, attrs) {
                // adds the tooltip to the body
                $scope.createTooltip = function (event) {
                    if (attrs.title || attrs.tooltip) {
                        var direction = $scope.getDirection();

                        // create the tooltip
                        $scope.tooltipElement = angular.element('<div>').addClass('angular-tooltip').addClass($scope.titleClass);

                        // append to the body
                        angular.element(document).find('body').append($scope.tooltipElement);

                        // update the contents and position
                        $scope.updateTooltip(attrs.title || attrs.tooltip);

                        // fade in
                        $scope.tooltipElement.addClass('angular-tooltip-fade-in');
                    }
                };

                $scope.updateTooltip = function(title) {
                    // insert html into tooltip
                    $scope.tooltipElement.html(title);

                    // compile html contents into angularjs
                    $compile($scope.tooltipElement.contents())($scope);

                    // calculate the position of the tooltip
                    var pos = $scope.calculatePosition($scope.tooltipElement, $scope.getDirection());
                    $scope.tooltipElement.addClass('angular-tooltip-' + pos.direction).css(pos);

                    // stop the standard tooltip from being shown
                    $timeout(function () {
                        element.removeAttr('ng-attr-title');
                        element.removeAttr('title');
                    });
                };

                // if the title changes the update the tooltip
                $scope.$watch('title', function(newTitle) {
                    if ($scope.tooltipElement) {
                        $scope.updateTooltip(newTitle);
                    }
                });

                // removes all tooltips from the document to reduce ghosts
                $scope.removeTooltip = function () {
                    var tooltip = angular.element(document.querySelectorAll('.angular-tooltip'));
                    // tooltip.removeClass('angular-tooltip-fade-in');

                    // $timeout(function() {
                        tooltip.remove();
                    // }, 300);
                };

                // gets the current direction value
                $scope.getDirection = function() {
                    return element.attr('tooltip-direction') || element.attr('title-direction') || 'top';
                };

                // calculates the position of the tooltip
                $scope.calculatePosition = function(tooltip, direction) {
                    var tooltipBounding = tooltip[0].getBoundingClientRect();
                    var elBounding = element[0].getBoundingClientRect();
                    var scrollLeft = window.scrollX || document.documentElement.scrollLeft;
                    var scrollTop = window.scrollY || document.documentElement.scrollTop;
                    var arrow_padding = 12;
                    var pos = {};
                    var newDirection = null;

                    // calculate the left position
                    if ($scope.stringStartsWith(direction, 'left')) {
                        pos.left = elBounding.left - tooltipBounding.width - (arrow_padding / 2) + scrollLeft;
                    } else if ($scope.stringStartsWith(direction, 'right')) {
                        pos.left = elBounding.left + elBounding.width + (arrow_padding / 2) + scrollLeft;
                    } else if ($scope.stringContains(direction, 'left')) {
                        pos.left = elBounding.left - tooltipBounding.width + arrow_padding + scrollLeft;
                    } else if ($scope.stringContains(direction, 'right')) {
                        pos.left = elBounding.left + elBounding.width - arrow_padding + scrollLeft;
                    } else {
                        pos.left = elBounding.left + (elBounding.width / 2) - (tooltipBounding.width / 2) + scrollLeft;
                    }

                    // calculate the top position
                    if ($scope.stringStartsWith(direction, 'top')) {
                        pos.top = elBounding.top - tooltipBounding.height - (arrow_padding / 2) + scrollTop;
                    } else if ($scope.stringStartsWith(direction, 'bottom')) {
                        pos.top = elBounding.top + elBounding.height + (arrow_padding / 2) + scrollTop;
                    } else if ($scope.stringContains(direction, 'top')) {
                        pos.top = elBounding.top - tooltipBounding.height + arrow_padding + scrollTop;
                    } else if ($scope.stringContains(direction, 'bottom')) {
                        pos.top = elBounding.top + elBounding.height - arrow_padding + scrollTop;
                    } else {
                        pos.top = elBounding.top + (elBounding.height / 2) - (tooltipBounding.height / 2) + scrollTop;
                    }

                    // check if the tooltip is outside the bounds of the window
                    if ($scope.fixedPosition) {
                        if (pos.left < scrollLeft) {
                            newDirection = direction.replace('left', 'right');
                        } else if ((pos.left + tooltipBounding.width) > (window.innerWidth + scrollLeft)) {
                            newDirection = direction.replace('right', 'left');
                        }

                        if (pos.top < scrollTop) {
                            newDirection = direction.replace('top', 'bottom');
                        } else if ((pos.top + tooltipBounding.height) > (window.innerHeight + scrollTop)) {
                            newDirection = direction.replace('bottom', 'top');
                        }

                        if (newDirection) {
                            return $scope.calculatePosition(tooltip, newDirection);
                        }
                    }

                    pos.left += 'px';
                    pos.top += 'px';
                    pos.direction = direction;

                    return pos;
                };

                $scope.stringStartsWith = function(searchString, findString) {
                    return searchString.substr(0, findString.length) === findString;
                };

                $scope.stringContains = function(searchString, findString) {
                    return searchString.indexOf(findString) !== -1;
                };

                if (attrs.title || attrs.tooltip) {
                    // attach events to show tooltip
                    element.on('mouseover', $scope.createTooltip);
                    element.on('mouseout', $scope.removeTooltip);
                } else {
                    // remove events
                    element.off('mouseover', $scope.createTooltip);
                    element.off('mouseout', $scope.removeTooltip);
                }

                element.on('destroy', $scope.removeTooltip);
                $scope.$on('$destroy', $scope.removeTooltip);
            }
        };
    };

    directive.$inject = ['$timeout', '$compile'];

    angular
        .module('tooltips', [])
        .directive('title', directive)
        .directive('tooltip', directive);
})();
