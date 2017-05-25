(function() {
    'use strict';

    var directive = function ($timeout, $compile) {
        return {
            restrict: 'A',
            scope: {
                title: '@'
            },
            link: function ($scope, element, attrs) {
                // adds the tooltip to the body
                $scope.createTooltip = function (event) {
                    if (attrs.title || attrs.tooltip) {
                        var direction = $scope.getDirection();

                        // create the tooltip
                        $scope.tooltipElement = angular.element('<div>')
                        .addClass('angular-tooltip angular-tooltip-' + direction);

                        // append to the body
                        angular.element(document).find('body').append($scope.tooltipElement);

                        $scope.updateTooltip(attrs.title || attrs.tooltip);

                        // fade in
                        $scope.tooltipElement.addClass('angular-tooltip-fade-in');
                    }
                };

                $scope.updateTooltip = function(title) {
                    $scope.tooltipElement.html(title);

                    $compile($scope.tooltipElement.contents())($scope);

                    var css = $scope.calculatePosition($scope.tooltipElement, $scope.getDirection());
                    $scope.tooltipElement.css(css);

                    // stop the standard tooltip from being shown
                    $timeout(function () {
                        element.removeAttr('ng-attr-title');
                        element.removeAttr('title');
                    });
                };

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

                $scope.calculatePosition = function(tooltip, direction) {
                    var tooltipBounding = tooltip[0].getBoundingClientRect();
                    var elBounding = element[0].getBoundingClientRect();
                    var scrollLeft = window.scrollX || document.documentElement.scrollLeft;
                    var scrollTop = window.scrollY || document.documentElement.scrollTop;
                    var arrow_padding = 12;

                    switch (direction) {
                        case 'top':
                        case 'top-center':
                        case 'top-middle':
                            return {
                                left: elBounding.left + (elBounding.width / 2) - (tooltipBounding.width / 2) + scrollLeft + 'px',
                                top: elBounding.top - tooltipBounding.height - (arrow_padding / 2) + scrollTop + 'px',
                            };
                        case 'top-right':
                            return {
                                left: elBounding.left + elBounding.width - arrow_padding + scrollLeft + 'px',
                                top: elBounding.top - tooltipBounding.height - (arrow_padding / 2) + scrollTop + 'px',
                            };
                        case 'right-top':
                            return {
                                left: elBounding.left + elBounding.width + (arrow_padding / 2) + scrollLeft + 'px',
                                top: elBounding.top - tooltipBounding.height + arrow_padding + scrollTop + 'px',
                            };
                        case 'right':
                        case 'right-center':
                        case 'right-middle':
                            return {
                                left: elBounding.left + elBounding.width + (arrow_padding / 2) + scrollLeft + 'px',
                                top: elBounding.top + (elBounding.height / 2) - (tooltipBounding.height / 2) + scrollTop + 'px',
                            };
                        case 'right-bottom':
                            return {
                                left: elBounding.left + elBounding.width + (arrow_padding / 2) + scrollLeft + 'px',
                                top: elBounding.top + elBounding.height - arrow_padding + scrollTop + 'px',
                            };
                        case 'bottom-right':
                            return {
                                left: elBounding.left + elBounding.width - arrow_padding + scrollLeft + 'px',
                                top: elBounding.top + elBounding.height + (arrow_padding / 2) + scrollTop + 'px',
                            };
                        case 'bottom':
                        case 'bottom-center':
                        case 'bottom-middle':
                            return {
                                left: elBounding.left + (elBounding.width / 2) - (tooltipBounding.width / 2) + scrollLeft + 'px',
                                top: elBounding.top + elBounding.height + (arrow_padding / 2) + scrollTop + 'px',
                            };
                        case 'bottom-left':
                            return {
                                left: elBounding.left - tooltipBounding.width + arrow_padding + scrollLeft + 'px',
                                top: elBounding.top + elBounding.height + (arrow_padding / 2) + scrollTop + 'px',
                            };
                        case 'left-bottom':
                            return {
                                left: elBounding.left - tooltipBounding.width - (arrow_padding / 2) + scrollLeft + 'px',
                                top: elBounding.top + elBounding.height - arrow_padding + scrollTop + 'px',
                            };
                        case 'left':
                        case 'left-center':
                        case 'left-middle':
                            return {
                                left: elBounding.left - tooltipBounding.width - (arrow_padding / 2) + scrollLeft + 'px',
                                top: elBounding.top + (elBounding.height / 2) - (tooltipBounding.height / 2) + scrollTop + 'px',
                            };
                        case 'left-top':
                            return {
                                left: elBounding.left - tooltipBounding.width - (arrow_padding / 2) + scrollLeft + 'px',
                                top: elBounding.top - tooltipBounding.height + arrow_padding + scrollTop + 'px',
                            };
                        case 'top-left':
                            return {
                                left: elBounding.left - tooltipBounding.width + arrow_padding + scrollLeft + 'px',
                                top: elBounding.top - tooltipBounding.height - (arrow_padding / 2) + scrollTop + 'px',
                            };
                    }
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
