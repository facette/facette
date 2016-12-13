angular.module('facette.ui.graph', [])

.directive('graph', function() {
    return {
        restrict: 'E',
        replace: true,
        scope: {
            graphIndex: '@',
            graphId: '@',
            graphData: '=?',
            graphOptions: '=?',
            graphAttrs: '=?'
        },
        controller: 'GraphController',
        templateUrl: 'templates/graph.html'
    };
})

.controller('GraphController', function($scope, $rootScope, $element, $pageVisibility, $timeout, $window, plots) {
    $scope.graph = null;

    $scope.options = {
        controls: true
    };

    if ($scope.graphOptions) {
        angular.extend($scope.options, $scope.graphOptions);
    }

    $scope.optionsRef = angular.copy($scope.options);

    $scope.loading = false;
    $scope.empty = false;
    $scope.error = false;
    $scope.modified = false;
    $scope.paused = false;
    $scope.timeout = null;
    $scope.refreshInterval = 0;
    $scope.stepActive = null;
    $scope.folded = typeof $scope.options.folded == 'boolean' ? $scope.options.folded : false;
    $scope.legendActive = false;
    $scope.zooming = false;

    var elementLeft = $element.offset().left,
        elementWidth = $element.width();

    function applyOptions(options) {
        angular.extend($scope.options, options);
    }

    function draw() {
        if (!$scope.data) {
            return;
        }

        var element = $element.children('.graph-container')[0],
            startTime = moment($scope.data.start),
            endTime = moment($scope.data.end);

        var chartCfg = {
            axis: {
                x: {
                    min: startTime.toDate(),
                    max: endTime.toDate(),
                    tick: {
                        count: Math.max(Math.floor($element.width() / 80), 2)
                    }
                },
                y: {
                    legend: $scope.data.options.yaxis_legend || null,
                    tick: {
                        count: 3
                    }
                }
            },
            bindTo: element,
            padding: graphPadding,
            series: [],
            constants: $scope.data.options.constants || [],
            stack: $scope.data.options.stack_mode,
            title: $scope.data.options && $scope.data.options.title ? $scope.data.options.title : null,
            subtitle: startTime.format(timeFormatDisplay) + ' — ' + endTime.format(timeFormatDisplay),
            tooltip: {
                date: {
                    format: function(date) {
                        return moment(date).format(timeFormatDisplay);
                    }
                }
            },
            type: $scope.data.options.type,
            zoom: {
                enabled: true,
                onStart: function() {
                    $scope.zooming = true;
                    $scope.$apply();
                },
                onSelect: function(start, end) {
                    var startTime = moment(start);

                    angular.extend($scope.options, {
                        time: startTime.format(timeFormatRFC3339),
                        range: timeToRange(moment(end).diff(startTime))
                    });

                    $scope.zooming = false;
                    $scope.$apply();
                }
            },
            events: {
                cursorMove: function(time) {
                    $rootScope.$emit('PropagateCursorPosition', time, $scope);
                }
            }
        };

        // Define unit formatter
        switch ($scope.data.options.yaxis_unit) {
        case graphYAxisUnitMetric:
            chartCfg.axis.y.tick.format = d3.format('.2s');
            break;

        case graphYAxisUnitBinary:
            chartCfg.axis.y.tick.format = function(value) {
                return formatSize(value);
            };
            break;

        default:
            chartCfg.axis.y.tick.format = d3.format('.2r');
        }

        // Append series to chart
        angular.forEach($scope.data.series, function(series) {
            var entry = {
                name: series.name,
                plots: series.plots,
                summary: series.summary
            };

            if (series.options && series.options.color) {
                entry.color = series.options.color;
            }

            chartCfg.series.push(entry);
        });

        // Reset element position and width
        elementLeft = undefined;
        elementWidth = undefined;

        try {
            $scope.chart = chart[$scope.chart ? 'update' : 'create'](chartCfg);
            $scope.$parent.$emit('GraphLoaded', $scope.graphIndex, $scope.graphId);
        } catch (e) {
            console.error('Failed to render graph: ' + e.name + (e.message ? ': ' + e.message : ''));
        }
    }

    function fetchData() {
        if ($scope.paused || $scope.folded) {
            return;
        } else if (!$scope.inView || !$rootScope.hasFocus) {
            $scope.deferred = true;
            return;
        }

        $scope.loading = true;
        $scope.empty = false;
        $scope.error = false;
        $scope.summary = {};

        var query = {
            range: $scope.options.range
        };

        if ($scope.options.time) {
            query.time = $scope.options.time;
        }

        if ($scope.graphId) {
            query.id = $scope.graphId;
        } else if ($scope.graphData) {
            query.graph = $scope.graphData;

            // Set range and sample values with graph options ones if any
            if (query.graph.options) {
                if (query.graph.options.range) {
                    query.range = query.graph.options.range;
                }

                if (query.graph.options.sample) {
                    query.sample = query.graph.options.sample;
                }
            }
        } else {
            $scope.loading = false;
            $scope.empty = true;
            return;
        }

        // Append attributes to request if any (used for collections templates)
        if ($scope.graphAttrs) {
            query.attributes = $scope.graphAttrs;
        }

        // Cancel previous refresh timeout if any
        if ($scope.timeout) {
            $timeout.cancel($scope.timeout);
            $scope.timeout = null;
        }

        // Fetch plots data
        plots.fetch(query, function(data) {
            // Apply options defaults
            data.options = angular.extend({
                type: graphTypeArea,
                stack_mode: null,
                yaxis_unit: graphYAxisUnitFixed
            }, data.options);

            // Draw graph
            $scope.data = data;
            $scope.loading = false;

            // Remove old rendered graph
            $element.find('.graph-container svg').remove();

            draw();

            // Register next draw if refresh interval set
            if ($scope.options.refresh_interval || data.options.refresh_interval) {
                $scope.refreshInterval = $scope.options.refresh_interval || data.options.refresh_interval;
                registerNextDraw();
            } else {
                $scope.refreshInterval = 0;
            }
        }, function() {
            $scope.data = null;
            $scope.loading = false;
            $scope.error = true;
        });
    }

    function emitChange() {
        $scope.$parent.$emit('GraphChanged', $scope.graphIndex, $scope.graphId, {
            folded: $scope.folded,
            legendActive: $scope.legendActive
        });
    }

    function resetLink() {
        if (!$scope.exportLink) {
            return;
        }

        $scope.exportLink
            .removeAttr('download')
            .removeAttr('href');
    }

    function registerNextDraw() {
        if (!$scope.refreshInterval) {
            return;
        }

        // Cancel previous refresh timeout if any
        if ($scope.timeout) {
            $timeout.cancel($scope.timeout);
            $scope.timeout = null;
        }

        // Register next draw
        $scope.timeout = $timeout(fetchData, $scope.refreshInterval * 1000);
    }

    // Define scope functions
    $scope.export = function(e) {
        if (!$scope.chart) {
            return;
        }

        $scope.exportLink = angular.element(e.target).closest('a');
        if ($scope.exportLink.attr('href')) {
            return;
        }

        var data = $scope.chart.getSVG(),
            canvas = document.createElement('canvas'),
            context;

        canvas.setAttribute('width', $scope.chart.svg.attr('width'));
        canvas.setAttribute('height', $scope.chart.svg.attr('height'));

        if (!canvas.getContext || !(context = canvas.getContext('2d'))) {
            console.error('Your browser doesn’t support mandatory Canvas feature');
            return;
        }

        var image = new Image();
        image.onload = function() {
            var name = slugify($scope.chart.config.title) +
                '_' + moment($scope.data.start).format(timeFormatFilename) +
                '_' + moment($scope.data.end).format(timeFormatFilename) +
                '.png';

            context.drawImage(image, 0, 0);

            $scope.exportLink
                .attr('download', name)
                .attr('href', canvas.toDataURL('image/png').replace('image/png', 'image/octet-stream'))
                .get(0).click();
        };
        image.src = 'data:image/svg+xml;base64,' + btoa(unescape(encodeURIComponent(data)));
    };

    $scope.moveStep = function(forward) {
        forward = typeof forward == 'boolean' ? forward : false;

        var startTime = moment($scope.data.start),
            delta = moment($scope.data.end).diff(startTime) / 4;

        if (!forward) {
            delta *= -1;
        }

        angular.extend($scope.options, {
            time: moment(startTime).add(delta).format(timeFormatRFC3339),
            range: ($scope.options.range || defaultTimeRange).replace(/^-/, '')
        });
    };

    $scope.propagate = function() {
        $rootScope.$emit('ApplyGraphOptions', {
            time: $scope.options.time,
            range: $scope.options.range
        });
    };

    $scope.reset = function() {
        $scope.options = angular.copy($scope.optionsRef);
    };

    $scope.refresh = function() {
        fetchData();
    };

    $scope.setRange = function(range) {
        if (range != 'custom') {
            angular.extend($scope.options, {range: '-' + range});
            return;
        }

        $rootScope.$emit('PromptTimeRange', function(time, range) {
            $scope.time = time;
            $scope.range = range;

            applyOptions({time: time, range: range});
        });
    };

    $scope.toggleFold = function(state) {
        $scope.folded = state;

        if (!state) {
            fetchData();
        } else {
            $scope.data = {};
        }

        emitChange();
    };

    $scope.toggleLegend = function(state) {
        if (!$scope.chart) {
            return;
        }

        $scope.legendActive = state;
        $scope.chart.toggleLegend(state);

        // Reset export link
        resetLink();

        emitChange();
    };

    $scope.zoom = function(zoomIn) {
        zoomIn = typeof zoomIn == 'boolean' ? zoomIn : true;

        var startTime = moment($scope.data.start),
            delta = moment($scope.data.end).diff(startTime),
            range;

        if (zoomIn) {
            range = timeToRange(delta / 2);
            delta /= 4;
        } else {
            range = timeToRange(delta * 2);
            delta = (delta / 2) * -1;
        }

        angular.extend($scope.options, {
            time: moment(startTime).add(delta).format(timeFormatRFC3339),
            range: range
        });
    };

    $scope.handleView = function(inView, info) {
        $scope.inView = inView;

        if (inView && info.changed && $scope.deferred) {
            $scope.deferred = false;
            fetchData();
        }
    };

    // Register watchers
    function updateGraph(newValue, oldValue) {
        if (angular.equals(newValue, oldValue)) {
            return;
        }

        // Check if options have been modified
        $scope.modified = !angular.equals($scope.options, $scope.optionsRef);

        // Reset export link
        resetLink();

        fetchData();
    }

    $scope.$watch('graphId', updateGraph, true);
    $scope.$watch('options', updateGraph, true);
    $scope.$watch('graphData', updateGraph, true);

    // Attach events
    $scope.$on('$destroy', function() {
        // Cancel existing refresh if any
        if ($scope.timeout) {
            $timeout.cancel($scope.timeout);
            $scope.timeout = null;
        }
    });

    $rootScope.$on('ApplyGraphOptions', function(e, options) {
        applyOptions(options);
    });

    $rootScope.$on('PropagateCursorPosition', function(e, time, origScope) {
        if (!$scope.chart || $scope === origScope) {
            return;
        }

        $scope.chart.toggleCursor(time);
    });

    $rootScope.$on('RedrawGraph', function(e) {
        draw();
    });

    $rootScope.$on('PauseGraphDraw', function(e, id, state) {
        if (id !== $scope.graphId) {
            return;
        }

        // Replace or register next draw that might be expired
        if (!state) {
            registerNextDraw();
        }

        $scope.paused = state;
    });

    $rootScope.$on('ToggleGraphLegend', function(e, idx, id, state) {
        if (idx && idx !== $scope.graphIndex || id && id !== $scope.graphId) {
            return;
        }

        $scope.toggleLegend(state);
    });

    $pageVisibility.$on('pageFocused', function(e) {
        if ($scope.deferred) {
            $scope.deferred = false;
            fetchData();
        }
    });

    angular.element($window).on('resize', function() {
        if ($scope.resizeTimeout) {
            $timeout.cancel($scope.resizeTimeout);
            $scope.resizeTimeout = null;
        }

        $scope.resizeTimeout = $timeout(draw, 50);
    });

    $element.on('mousemove', function(e) {
        if ($scope.zooming) {
            return;
        }

        if (elementLeft === undefined || elementWidth === undefined) {
            elementLeft = $element.offset().left;
            elementWidth = $element.width();
        }

        var changed = false;

        if (e.pageX - elementLeft <= graphPadding * 2) {
            $scope.stepActive = 'backward';
            changed = true;
        } else if (e.pageX - elementLeft >= elementWidth - graphPadding * 2) {
            $scope.stepActive = 'forward';
            changed = true;
        } else if ($scope.stepActive !== null) {
            $scope.stepActive = null;
            changed = true;
        }

        if (changed) {
            $scope.$apply();
        }
    });

    // Set range values
    $scope.rangeValues = timeRanges;

    // Trigger first draw
    fetchData();
});
