angular.module('facette.ui.graph', [])

.directive('graph', function() {
    return {
        restrict: 'E',
        replace: true,
        scope: {
            index: '@',
            graphId: '@',
            def: '=?',
            options: '=?',
            attributes: '=?',
            controls: '@',
            frame: '@'
        },
        link: function(scope, element, attrs) {
            attrs.$observe('controls', function() { scope.controls = scope.$eval(attrs.controls); });
            attrs.$observe('frame', function() { scope.frame = scope.$eval(attrs.frame); });
        },
        controller: 'GraphController',
        templateUrl: 'templates/graph.html'
    };
})

.controller('GraphController', function($scope, $rootScope, $element, $pageVisibility, $timeout, $window, series) {
    $scope.graph = null;

    if (!angular.isDefined($scope.options)) {
        $scope.options = {};
    }
    $scope.optionsRef = angular.copy($scope.options);

    $scope.embeddablePath = $scope.options.embeddable_path || null;

    $scope.startTime = null;
    $scope.endTime = null;
    $scope.time = null;
    $scope.range = null;

    $scope.loading = false;
    $scope.empty = false;
    $scope.partial = false;
    $scope.error = false;
    $scope.modified = false;
    $scope.paused = false;
    $scope.timeout = null;
    $scope.refreshInterval = 0;
    $scope.stepActive = null;
    $scope.folded = typeof $scope.options.folded == 'boolean' ? $scope.options.folded : false;
    $scope.legendActive = false;
    $scope.zooming = false;
    $scope.exportLinks = {};

    var elementLeft = $element.offset().left,
        elementTop = $element.offset().top,
        elementWidth = $element.width();

    function applyOptions(options, force) {
        force = typeof force == 'boolean' ? force : false;

        var optionsOrig = angular.copy($scope.options),
            optionsNew = angular.copy($scope.options),
            embeddablePath = null;

        angular.extend(optionsNew, options);

        if (options.start_time || options.end_time) {
            delete optionsNew.time;
            delete optionsNew.range;

            if ($scope.options.embeddable_path) {
                embeddablePath = $scope.options.embeddable_path +
                    "?start=" + encodeURIComponent(moment(options.start_time).format(timeFormatRFC3339)) +
                    "&end=" + encodeURIComponent(moment(options.end_time).format(timeFormatRFC3339));
            }
        } else if (options.time || options.range) {
            delete optionsNew.start_time;
            delete optionsNew.end_time;

            if ($scope.options.embeddable_path) {
                embeddablePath = $scope.options.embeddable_path + "?";

                if (options.time) {
                    embeddablePath += "time=" + encodeURIComponent(moment(options.time).format(timeFormatRFC3339));
                }

                if (options.range) {
                    embeddablePath += (options.time ? '&' : '') + "range=" + options.range;
                }
            }
        } else {
            delete optionsNew.time;
            delete optionsNew.range;
            delete optionsNew.start_time;
            delete optionsNew.end_time;

            if ($scope.options.embeddable_path) {
                embeddablePath = $scope.options.embeddable_path;
            }
        }

        if (embeddablePath) {
            $scope.embeddablePath = embeddablePath;
        }

        $scope.options = optionsNew;

        if (angular.equals(optionsNew, optionsOrig) && force) {
            updateGraph(optionsNew, null);
        }
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
                    label: $scope.data.options.yaxis_label || null,
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

                    applyOptions({
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

        // Set Y-Axis extremes and centering
        if ($scope.data.options.yaxis_min) {
            chartCfg.axis.y.min = $scope.data.options.yaxis_min;
        }

        if ($scope.data.options.yaxis_max) {
            chartCfg.axis.y.max = $scope.data.options.yaxis_max;
        }

        if (typeof $scope.data.options.yaxis_center == 'boolean') {
            chartCfg.axis.y.center = $scope.data.options.yaxis_center;
        }

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
            if (series.points === null) {
                $scope.partial = true;
            }

            var entry = {
                name: series.name,
                points: series.points,
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
            $scope.$parent.$emit('GraphLoaded', $scope.index, $scope.graphId);
        } catch (e) {
            console.error('Failed to render graph: ' + e.name + (e.message ? ': ' + e.message : ''));
        }
    }

    function fetchData() {
        if ($scope.paused || $scope.folded) {
            return;
        }

        if (!$scope.inView || !$rootScope.hasFocus) {
            $scope.deferred = true;
            return;
        }

        $scope.loading = true;
        $scope.empty = false;
        $scope.partial = false;
        $scope.error = false;
        $scope.summary = {};

        var query = {
            normalize: 1
        };

        angular.forEach(['start_time', 'end_time', 'time', 'range'], function(key) {
            if ($scope.options[key]) {
                query[key] = $scope.options[key];
            }
        });

        if ($scope.graphId) {
            query.id = $scope.graphId;
        } else if ($scope.def) {
            query.graph = $scope.def;

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
        if ($scope.attributes) {
            query.attributes = $scope.attributes;
        }

        // Cancel previous refresh timeout if any
        if ($scope.timeout) {
            $timeout.cancel($scope.timeout);
            $scope.timeout = null;
        }

        // Fetch data points
        series.points(query, function(data) {
            // Apply options defaults
            data.options = angular.extend({
                type: graphTypeArea,
                stack_mode: null,
                yaxis_unit: graphYAxisUnitFixed
            }, data.options);

            // Draw graph
            $scope.data = data;
            $scope.loading = false;

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
            $scope.partial = false;
            $scope.error = true;

            // Remove old rendered graph
            $element.find('.graph-container svg').remove();
        });
    }

    function emitChange() {
        $scope.$parent.$emit('GraphChanged', $scope.index, $scope.graphId, {
            folded: $scope.folded,
            legendActive: $scope.legendActive
        });
    }

    function resetLink() {
        if (!$scope.exportLinks) {
            return;
        }

        angular.forEach($scope.exportLinks, function(link) {
            link.removeAttr('download').removeAttr('href');
        });
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
    $scope.export = function(e, type) {
        if (!$scope.chart) {
            return;
        }

        $scope.exportLinks[type] = angular.element(e.target).closest('a');
        if ($scope.exportLinks[type].attr('href')) {
            return;
        }

        switch (type) {
        case 'png':
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

                $scope.exportLinks[type]
                    .attr('download', name)
                    .attr('href', canvas.toDataURL('image/png').replace('image/png', 'image/octet-stream'))
                    .get(0).click();
            };

            image.src = 'data:image/svg+xml;base64,' + btoa(unescape(encodeURIComponent(data)));

            break;

        case 'summary_csv':
        case 'summary_json':
            var name = slugify($scope.chart.config.title) +
                '_' + moment($scope.data.start).format(timeFormatFilename) +
                '_' + moment($scope.data.end).format(timeFormatFilename) +
                '_' + type.replace('_', '.');

            var hrefData,
                summary;

            if (type == 'summary_csv') {
                summary = '';
                angular.forEach($scope.data.series, function(series, idx) {
                    var keys = Object.keys(series.summary);

                    if (idx === 0) {
                        summary += 'name,' + keys.join(',') + '\n';
                    }

                    summary += '"' + series.name + '",' +
                        keys.map(function(x) { return series.summary[x]; }).join(',') + '\n';
                });

                hrefData = 'data:text/csv;charset=utf-8,' + encodeURIComponent(summary);
            } else {
                summary = {};
                angular.forEach($scope.data.series, function(series) {
                    summary[series.name] = series.summary;
                });

                hrefData = 'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(summary, null, '\t'));
            }

            $timeout(function() {
                $scope.exportLinks[type]
                    .attr('download', name)
                    .attr('href', hrefData)
                    .get(0).click();
            }, 0);

            break;
        }
    };

    $scope.moveStep = function(forward) {
        forward = typeof forward == 'boolean' ? forward : false;

        var endTime = moment($scope.data.end),
            delta = moment($scope.data.start).diff(endTime) / 4;

        if (forward) {
            delta *= -1;
        }

        applyOptions({
            time: moment(endTime).add(delta).format(timeFormatRFC3339)
        });
    };

    $scope.propagate = function() {
        var options = {};

        if ($scope.options.start_time || $scope.options.end_time) {
            options.start_time = $scope.options.start_time;
            options.end_time = $scope.options.end_time;
        } else if ($scope.options.time || $scope.options.range) {
            options.time = $scope.options.time;
            options.range = $scope.options.range;
        } else {
            options.start_time = options.end_time = options.time = options.range = null;
        }

        $rootScope.$emit('ApplyGraphOptions', options, true);
    };

    $scope.reset = function() {
        $scope.options = angular.copy($scope.optionsRef);
    };

    $scope.refresh = function() {
        fetchData();
    };

    $scope.setRange = function(range) {
        if (range != 'custom') {
            applyOptions({range: '-' + range});
            return;
        }

        $rootScope.$emit('PromptTimeRange', function(startTime, endTime, time, range) {
            $scope.startTime = startTime;
            $scope.endTime = endTime;
            $scope.time = time;
            $scope.range = range;

            applyOptions(startTime && endTime ? {
                start_time: startTime,
                end_time: endTime
            } : {
                time: time,
                range: range
            });
        }, {
            start: $scope.startTime,
            end: $scope.endTime,
            time: $scope.time,
            range: $scope.range
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

        applyOptions({
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
    $scope.$watch('def', updateGraph, true);

    // Attach events
    var unregisterCallbacks = [];

    unregisterCallbacks.push($rootScope.$on('ApplyGraphOptions', function(e, options, force) {
        applyOptions(options, force);
    }));

    unregisterCallbacks.push($rootScope.$on('PropagateCursorPosition', function(e, time, origScope) {
        if (!$scope.chart || $scope === origScope) {
            return;
        }

        $scope.chart.toggleCursor(time);
    }));

    unregisterCallbacks.push($rootScope.$on('ResetTimeRange', function() {
        $scope.reset();
    }));

    unregisterCallbacks.push($rootScope.$on('RedrawGraph', function() {
        draw();
    }));

    unregisterCallbacks.push($rootScope.$on('RefreshGraph', function() {
        fetchData();
    }));

    unregisterCallbacks.push($rootScope.$on('PauseGraphDraw', function(e, id, state) {
        if (id !== $scope.graphId) {
            return;
        }

        // Replace or register next draw that might be expired
        if (!state) {
            registerNextDraw();
        }

        $scope.paused = state;
    }));

    unregisterCallbacks.push($rootScope.$on('ToggleGraphLegend', function(e, idx, id, state) {
        if (idx && idx !== $scope.index || id && id !== $scope.graphId) {
            return;
        }

        $scope.toggleLegend(state);
    }));

    $scope.$on('$destroy', function() {
        // Cancel existing refresh if any
        if ($scope.timeout) {
            $timeout.cancel($scope.timeout);
            $scope.timeout = null;
        }

        angular.forEach(unregisterCallbacks, function(callback) {
            callback();
        });
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

        if (elementLeft === undefined || elementTop === undefined || elementWidth === undefined) {
            var elementOffset = $element.offset();

            elementLeft = elementOffset.left;
            elementTop = elementOffset.top;
            elementWidth = $element.width();
        }

        var changed = false,
            relX = e.pageX - elementLeft,
            relY = e.pageY - elementTop,
            delta = graphPadding * 2;

        if (!$scope.stepActive && relX <= delta) {
            $scope.stepActive = 'backward';
            changed = true;
        } else if (!$scope.stepActive && relX >= elementWidth - delta) {
            $scope.stepActive = 'forward';
            changed = true;
        } else if ($scope.stepActive !== null && relX > delta && relX < elementWidth - delta) {
            $scope.stepActive = null;
            changed = true;
        }

        if (!$scope.foldActive && relY <= delta) {
            $scope.foldActive = true;
            changed = true;
        } else if ($scope.foldActive && relY > delta) {
            $scope.foldActive = false;
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
