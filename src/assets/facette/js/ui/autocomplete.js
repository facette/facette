angular.module('facette.ui.autocomplete', [])

.directive('autocomplete', function() {
    return {
        restrict: 'E',
        replace: true,
        scope: {
            id: '@',
            delay: '@',
            allowOverride: '=?',
            source: '=',
            onSelect: '=?'
        },
        controller: 'AutocompleteController',
        templateUrl: 'templates/autocomplete.html'
    };
})

.controller('AutocompleteController', function($element, $sce, $scope, $timeout) {
    $scope.selected = true;

    if (!angular.isDefined($scope.delay)) {
        $scope.delay = 250;
    }

    // Define scope functions
    $scope.activate = function(idx) {
        $scope.index = idx;
    };

    $scope.handleFocus = function(e) {
        $scope.focus = e.type == 'focus';

        if (e.type == 'blur') {
            if ($scope.allowOverride && !$scope.selected) {
                $scope.select(e, {label: e.target.value, value: e.target.value});
            } else {
                delete $scope.entries;
            }
        }
    };

    $scope.handleKey = function(e) {
        switch (e.which) {
        case 13: // <Enter>
            if ($scope.entries !== undefined) {
                $scope.select(e, $scope.entries[$scope.index]);
            } else if ($scope.allowOverride && !$scope.selected) {
                $scope.select(e, {label: e.target.value, value: e.target.value});
            }

            break;

        case 27: // <Escape>
            delete $scope.entries;

            break;

        case 38: // <Up>
        case 40: // <Down>
            e.preventDefault();

            if (e.which == 38 && $scope.index > 0) {
                $scope.index--;
            } else if (e.which == 40 && $scope.index < $scope.entries.length - 1) {
                $scope.index++;
            }

            // Update scroll position
            $timeout(function() {
                var row = $element.find('.autocomplete-row.active'),
                    position = row.position(),
                    container = row.parent();

                if (position.top + row.outerHeight(true) > container.innerHeight()) {
                    container.scrollTop(container.scrollTop() + position.top - container.innerHeight() +
                        row.outerHeight(true) + (container.outerHeight() - container.innerHeight()));
                } else if (position.top < 0) {
                    container.scrollTop(container.scrollTop() + position.top);
                }
            }, 0);

            break;
        }
    };

    $scope.highlight = function(input, value) {
        try {
            return $sce.trustAsHtml(input.replace(new RegExp('(' + value + ')', 'gi'), '<mark>$1</mark>'));
        } catch (e) {
            return input;
        }
    };

    $scope.select = function(e, data) {
        unwatchValue();
        $scope.value = data.label;
        watchValue();

        if (angular.isDefined($scope.onSelect)) {
            $scope.onSelect(e, data.value);
        }

        $scope.selected = true;

        delete $scope.entries;
    };

    // Register watchers
    var unwatchValue;

    function watchValue() {
        unwatchValue = $scope.$watch('value', function(newValue, oldValue) {
            if (!newValue) {
                delete $scope.entries;
                return;
            } else if (angular.equals(newValue, oldValue)) {
                return;
            }

            if ($scope.completeTimeout) {
                $timeout.cancel($scope.completeTimeout);
                $scope.completeTimeout = null;
            }

            $scope.completeTimeout = $timeout(function() {
                $scope.selected = false;
                $scope.activate(0);

                $scope.source(newValue).then(function(data) {
                    $scope.entries = data;
                });
            }, $scope.delay);
        });
    }

    watchValue();
});
