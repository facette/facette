app.factory('timeRange', function($timeout, ngDialog) {
    return {
        prompt: function(callback) {
            // Handle custom range selection
            ngDialog.open({
                template: 'templates/common/timerange.html',
                controller: function($scope) {
                    function resetDatetime() {
                        $scope.datetime = {
                            time: false,
                            start: false,
                            end: false
                        };
                    }

                    $scope.data = {
                        absolute: false
                    };

                    $scope.switchAbsolute = function(data, state) {
                        data.absolute = state;

                        $scope.$applyAsync(function() {
                            angular.element('.ngdialog-content :input:visible:first').select();
                        });
                    };

                    $scope.hideDatetime = function(e) {
                        if (angular.element(e.target).closest('.datetimepicker, .datetimepicker-holder').length === 0) {
                            resetDatetime();
                        }
                    };

                    $scope.toggleDatetime = function(name, state) {
                        angular.forEach(Object.keys($scope.datetime), function(entry) {
                            $scope.datetime[entry] = entry === name ? state : false;
                        });
                    };

                    resetDatetime();
                },
                showClose: false
            }).closePromise.then(function(scope) {
                scope.$dialog.remove();

                if ([undefined, '$document', '$escape'].indexOf(scope.value) != -1) {
                    return;
                }

                if (scope.value.absolute) {
                    callback(scope.value.start, timeToRange(moment(scope.value.end).diff(scope.value.start)));
                } else {
                    callback(scope.value.time, scope.value.range);
                }
            });
        }
    };
});
