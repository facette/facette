
/* Graph */

var GRAPH_DRAW_PARENTS  = [],
    GRAPH_DRAW_QUEUE    = [],
    GRAPH_DRAW_TIMEOUTS = {},

    GRAPH_CONTROL_LOCK    = false,
    GRAPH_CONTROL_TIMEOUT = null,

    $graphTemplate;

function graphDraw(graph, postpone, delay, preview) {
    var graphNew;

    postpone = typeof postpone == 'boolean' ? postpone : false;
    delay    = delay || 0;
    preview  = preview || null;

    if (graph.length > 1) {
        console.error("Can't draw multiple graph.");
        return;
    }

    if (!graph.data('setup')) {
        // Replace node with graph template
        if ($graphTemplate.length > 0) {
            graphNew = $graphTemplate.clone();

            $.each(graph.prop("attributes"), function () {
                graphNew.attr(this.name, this.value);
            });

            graph.replaceWith(graphNew);
            graph = graphNew;

            graph.data({
                options: graph.opts('graph'),
                setup: true
            });

            graph.find('.graphctrl .ranges').hide();

            if (readOnly)
                graph.find('.graphctrl .edit').hide();

            graph.find('.placeholder').text(graph.data('options').title || 'N/A');
        }
    }

    // Clear previous refresh timeout
    if (graph.data('timeout')) {
        clearTimeout(graph.data('timeout'));
        graph.removeData('timeout');
    }

    // Postpone graph draw
    if (postpone) {
        graphEnqueue(graph.get(0));
        return;
    }

    return $.Deferred(function ($deferred) {
        setTimeout(function () {
            var graphOpts,
                query,
                location,
                args;

            graph.find('.placeholder').text($.t('main.mesg_loading'));

            // Parse graph options
            graphOpts = graph.data('options') || graph.opts('graph');

            if (typeof graphOpts.zoom == 'undefined')
                graphOpts.zoom = true;

            if (typeof graphOpts.expand == 'undefined')
                graphOpts.expand = true;

            if (typeof graphOpts.legend == 'undefined')
                graphOpts.legend = false;

            if (graphOpts.sample)
                graphOpts.sample = parseInt(graphOpts.sample, 10);
            else
                delete graphOpts.sample;

            if (!graphOpts.range)
                graphOpts.range = GRAPH_DEFAULT_RANGE;

            if (typeof graphOpts.percentiles != 'undefined') {
                switch (typeof graphOpts.percentiles) {
                case 'number':
                    graphOpts.percentiles = [graphOpts.percentiles];
                    break;

                case 'string':
                    graphOpts.percentiles = parseFloatList(graphOpts.percentiles);
                    break;
                }
            }

            if (typeof graphOpts.constants != 'undefined') {
                switch (typeof graphOpts.constants) {
                case 'number':
                    graphOpts.constants = [graphOpts.constants];
                    break;

                case 'string':
                    graphOpts.constants = parseFloatList(graphOpts.constants);
                    break;
                }
            }

            // Update URL on show
            if (locationPath.startsWith(urlPrefix + '/show/')) {
                location = String(window.location.pathname);

                args = [];
                if (graphOpts.time)
                    args.push('time=' + graphOpts.time.replace('+', '%2B'));
                if (graphOpts.range)
                    args.push('range=' + graphOpts.range);
                if (graphOpts.refresh_interval)
                    args.push('refresh=' + graphOpts.refresh_interval);

                if (args.length > 0)
                    location += '?' + args.join('&');

                if (location != (window.location.pathname + window.location.search))
                    history.replaceState(null, document.title, location);
            }

            // Set graph options
            graph.data('options', graphOpts);

            // Render graph plots
            query = {
                time: graphOpts.time,
                range: graphOpts.range,
                sample: graphOpts.sample,
                percentiles: graphOpts.percentiles
            };

            if (preview) {
                query.graph = preview;

                graphOpts.legend = false;
            } else {
                query.id = graph.attr('data-graph');
            }

            return $.ajax({
                url: urlPrefix + '/api/v1/plots',
                type: 'POST',
                contentType: 'application/json',
                data: JSON.stringify(query),
                dataType: 'json'
            }).pipe(function (data) {
                var $container,
                    graphTableUpdate,
                    highchart,
                    highchartOpts,
                    startTime,
                    endTime,
                    seriesData = {},
                    seriesIndexes = [],
                    seriesVisibility = {},
                    seriesPlotlines = [],
                    i,
                    j;

                if (data.message || !data.series) {
                    graph.children('.graphctrl')
                        .attr('disabled', 'disabled')
                        .find('a:not([href="#edit"], [href="#refresh"], [href="#reset"]), .legend')
                            .attr('disabled', 'disabled');

                    graph.find('.placeholder')
                        .addClass('icon icon-warning')
                        .text(data.message ? data.message : $.t('graph.mesg_empty_series'))
                        .show();

                    graph.children('.graphcntr').empty();

                    $deferred.resolve();

                    return;
                } else {
                    graph.children('.graphctrl')
                        .removeAttr('disabled')
                        .find('a:not([href="#edit"], [href="#refresh"], [href="#reset"]), .legend')
                            .removeAttr('disabled');

                    graph.find('.placeholder')
                        .removeClass('icon icon-warning')
                        .hide();
                }

                startTime = moment(data.start);
                endTime   = moment(data.end);

                graphTableUpdate = function () {
                    if (graphOpts.legend)
                        Highcharts.drawTable.apply(this, [seriesData]);
                };

                highchartOpts = {
                    chart: {
                        borderRadius: 0,
                        events: {
                            load: graphTableUpdate,
                            redraw: graphTableUpdate,
                            togglePlotLine: function () {
                                var $element,
                                    regexp = new RegExp('(^| +)active( +|$)'),
                                    name;

                                $element = $(this.element);

                                name = 'plotline-' + this.series.name + '-' + this.name;

                                // Remove existing plot line
                                this.chart.yAxis[0].removePlotLine(name);

                                if ($element.attr('class') && $element.attr('class').match(regexp)) {
                                    $element.css({
                                        color: 'inherit',
                                        fill: 'inherit'
                                    }).attr('class', $element.attr('class').replace(regexp, ''));

                                    return;
                                }

                                // Set element active
                                if (!this.chart.options._data.plotlines[name])
                                    this.chart.options._data.plotlines[name] = GRAPH_PLOTLINE_COLORS[Object.keys(this
                                        .chart.options._data.plotlines).length % GRAPH_PLOTLINE_COLORS.length];

                                $element
                                    .css({
                                        color: this.chart.options._data.plotlines[name],
                                        fill: this.chart.options._data.plotlines[name]
                                    })
                                    .attr('class', $element.attr('class') + ' active');

                                // Draw new plot line
                                this.chart.yAxis[0].addPlotLine({
                                    id: name,
                                    color: this.chart.options._data.plotlines[name],
                                    value: this.value,
                                    width: 1.5,
                                    zIndex: 3
                                });
                            }
                        },
                        spacingBottom: GRAPH_SPACING_SIZE * 2,
                        spacingLeft: GRAPH_SPACING_SIZE,
                        spacingRight: GRAPH_SPACING_SIZE,
                        spacingTop: GRAPH_SPACING_SIZE,
                    },
                    credits: {
                        enabled: false
                    },
                    exporting: {
                        enabled: false
                    },
                    legend: {
                        enabled: false
                    },
                    plotOptions: {},
                    series: [],
                    title: {
                        text: null
                    },
                    tooltip: {
                        formatter: function () {
                            var tooltip = '<strong>' + moment(this.x).format(TIME_DISPLAY) + '</strong>',
                                stacks = {},
                                i,
                                stackName,
                                total;

                            for (i in this.points) {
                                if (!stacks[this.points[i].series.stackKey])
                                    stacks[this.points[i].series.stackKey] = [];

                                stacks[this.points[i].series.stackKey].push({
                                    name: this.points[i].series.name,
                                    value: this.points[i].y,
                                    color: this.points[i].series.color,
                                    symbol: getHighchartsSymbol(this.points[i].series.symbol)
                                });
                            }

                            for (stackName in stacks) {
                                tooltip += '<div class="highcharts-tooltip-block">';

                                total = 0;

                                for (i in stacks[stackName]) {
                                    tooltip += '<div><span style="color: ' + stacks[stackName][i].color + '">' +
                                        stacks[stackName][i].symbol +'</span> ' + stacks[stackName][i].name +
                                        ': <strong>' + (stacks[stackName][i].value !== null ?
                                        formatValue(stacks[stackName][i].value, {unit_type: data.unit_type}) : 'null') +
                                        '</strong></div>';

                                    if (stacks[stackName][i].value !== null)
                                        total += stacks[stackName][i].value;
                                }

                                if (stacks[stackName].length > 1) {
                                    tooltip += '<div class="highcharts-tooltip-total">Total: <strong>' +
                                        (total !== null ? formatValue(total, {unit_type: data.unit_type}) : 'null') +
                                        '</strong></div>';
                                }
                            }

                            return tooltip;
                        },
                        shared: true,
                        useHTML: true
                    },
                    xAxis: {
                        max: endTime.valueOf(),
                        min: startTime.valueOf(),
                        type: 'datetime'
                    },
                    yAxis: {
                        labels: {
                            formatter: function () {
                                return formatValue(this.value, {unit_type: data.unit_type});
                            }
                        },
                        plotLines: [],
                        title: {
                            text: null
                        }
                    },
                    _data: {
                        plotlines: {}
                    },
                    _opts: data
                };

                // Set type-specific options
                switch (data.type) {
                case GRAPH_TYPE_AREA:
                    highchartOpts.chart.type = 'area';
                    break;
                case GRAPH_TYPE_LINE:
                    highchartOpts.chart.type = 'line';
                    break;
                default:
                    console.error("Unknown `" + data.type + "' chart type");
                    break;
                }

                highchartOpts.plotOptions[highchartOpts.chart.type] = {
                    animation: false,
                    lineWidth: 1.5,
                    marker: {
                        enabled: false
                    },
                    states: {
                        hover: {
                            lineWidth: 2.5
                        }
                    },
                    threshold: 0
                };

                // Enable full features when not in preview
                if (preview) {
                    highchartOpts.plotOptions[highchartOpts.chart.type].enableMouseTracking = false;

                    graph.children('.graphctrl').remove();
                } else {
                    highchartOpts.title = {
                        text: data.title || data.name
                    };

                    highchartOpts.subtitle = {
                        text: startTime.format(TIME_DISPLAY) + ' — ' + endTime.format(TIME_DISPLAY)
                    };

                    if (data.unit_legend)
                        highchartOpts.yAxis.title.text = data.unit_legend;

                    if (graphOpts.zoom) {
                        highchartOpts.chart.events.selection = function (e) {
                            if (e.xAxis) {
                                graphUpdateOptions(graph, {
                                    time: moment(e.xAxis[0].min).format(TIME_RFC3339),
                                    range: timeToRange(moment.duration(moment(e.xAxis[0].max)
                                        .diff(moment(e.xAxis[0].min))))
                                });

                                graphDraw(graph);
                            }

                            e.preventDefault();
                        };

                        highchartOpts.chart.zoomType = 'x';
                    }
                }

                // Set stacking options
                switch (data.stack_mode) {
                case STACK_MODE_NORMAL:
                    highchartOpts.plotOptions[highchartOpts.chart.type].stacking = 'normal';
                    break;
                case STACK_MODE_PERCENT:
                    highchartOpts.plotOptions[highchartOpts.chart.type].stacking = 'percent';
                    break;
                default:
                    highchartOpts.plotOptions[highchartOpts.chart.type].stacking = null;
                    break;
                }

                // Check for previous series visibility
                $container = graph.children('.graphcntr');

                highchart = $container.highcharts();
                if (highchart) {
                    $.each(highchart.series, function () {
                        seriesVisibility[this.name] = this.visible;
                    });
                    $.each(highchart.yAxis[0].plotLinesAndBands, function () {
                        seriesPlotlines.push(this.id);
                    });
                }

                // Append series data
                seriesIndexes = graphGetSeriesIndexes(data.series);

                for (i in data.series) {
                    // Transform unix epochs to Date objects
                    for (j in data.series[i].plots)
                        data.series[i].plots[j] = [data.series[i].plots[j][0] * 1000, data.series[i].plots[j][1]];

                    highchartOpts.series.push({
                        id: data.series[i].name,
                        name: data.series[i].name,
                        stack: 'stack' + data.series[i].stack_id,
                        data: data.series[i].plots,
                        color: data.series[i].options ? data.series[i].options.color : null,
                        visible: typeof seriesVisibility[data.series[i].name] !== undefined ?
                            seriesVisibility[data.series[i].name] : true,
                        zIndex: seriesIndexes.indexOf(data.series[i].name)
                    });

                    seriesData[data.series[i].name] = {
                        summary: data.series[i].summary,
                        options: data.series[i].options
                    };
                }

                // Prepare legend spacing
                if (graphOpts.legend) {
                    highchartOpts.chart.spacingBottom = highchartOpts.series.length * GRAPH_LEGEND_ROW_HEIGHT +
                        highchartOpts.chart.spacingBottom;

                    if (graph.data('toggled-legend') && graphOpts.expand) {
                        $container.height($container.outerHeight() + highchartOpts.series.length *
                            GRAPH_LEGEND_ROW_HEIGHT);

                        graph.data('toggled-legend', false);
                    }
                } else {
                    highchartOpts.chart.spacingBottom = GRAPH_SPACING_SIZE * 2;

                    if (graph.data('toggled-legend') && graphOpts.expand) {
                        $container.height($container.outerHeight() - highchartOpts.series.length *
                            GRAPH_LEGEND_ROW_HEIGHT);

                        graph.data('toggled-legend', false);
                    }
                }

                highchart = $container.highcharts(highchartOpts).highcharts();

                // Draw constants plot lines
                for (i in graphOpts.constants) {
                    highchart.yAxis[0].addPlotLine({
                        color: '#d00',
                        value: graphOpts.constants[i],
                        width: 1,
                        zIndex: 3
                    });
                }

                // Re-apply plotlines if any
                if (seriesPlotlines.length > 0) {
                    $.each(seriesPlotlines, function(i, name) {
                        if (name.startsWith('plotline-'))
                            name = name.substr(9);

                        graph.find('.graphcntr .highcharts-table-value[data-name="' + name + '"]').trigger('click');
                    });
                }

                // Set next refresh if needed
                if (graphOpts.refresh_interval) {
                    graph.data('timeout', setTimeout(function () {
                        graphDraw(graph, !graph.inViewport());
                    }, graphOpts.refresh_interval * 1000));
                }

                $deferred.resolve();
            }).fail(function () {
                graph.children('.graphctrl')
                    .attr('disabled', 'disabled')
                    .find('a:not([href="#edit"], [href="#refresh"], [href="#reset"]), .legend')
                        .attr('disabled', 'disabled');

                graph.find('.placeholder')
                    .addClass('icon icon-warning')
                    .html($.t('graph.mesg_load_failed', {name: '<strong>' + (graphOpts.title ||
                        graph.attr('data-graph')) + '</strong>'}));

                $deferred.resolve();
            });
        }, delay);
    }).promise();
}

function graphEnqueue(graph) {
    var $parent = $(graph).offsetParent(),
        parent = $parent.get(0),
        index = GRAPH_DRAW_PARENTS.indexOf(parent);

    if (index == -1) {
        GRAPH_DRAW_PARENTS.push(parent);
        GRAPH_DRAW_QUEUE.push([]);
        index = GRAPH_DRAW_PARENTS.length - 1;

        $parent.on('scroll', graphHandleQueue);
    }

    if (GRAPH_DRAW_QUEUE[index].indexOf(graph) == -1)
        GRAPH_DRAW_QUEUE[index].push(graph);
}

function graphExport(graph) {
    var canvas = document.createElement('canvas'),
        svg = graph.find('.graphcntr').highcharts().getSVG();

    canvas.setAttribute('width', parseInt(svg.match(/width="([0-9]+)"/)[1], 10));
    canvas.setAttribute('height', parseInt(svg.match(/height="([0-9]+)"/)[1], 10));

    if (canvas.getContext && canvas.getContext('2d')) {
        canvg(canvas, svg);

        window.location.href = canvas.toDataURL('image/png')
            .replace('image/png', 'image/octet-stream');

    } else {
        console.error("Your browser doesn't support mandatory Canvas feature");
    }
}

function graphGetSeriesIndexes(series) {
    var ordered = series.slice(0),
        indexes = [];

    ordered.sort(function (a, b) {
        if (!a.summary || !b.summary || !a.summary.avg || !b.summary.avg)
            return 0;

        return b.summary.avg - a.summary.avg;
    });

    $.each(ordered, function (index, entry) {
        indexes.push(entry.name);
    });

    return indexes;
}

function graphHandleActions(e) {
    var $target = $(e.target),
        $graph = $target.closest('[data-graph]'),
        $overlay,
        graphObj,
        delta,
        location,
        args = [],
        options,
        range;

    if (e.target.getAttribute('disabled') == 'disabled') {
        e.preventDefault();
        return;
    }

    if (e.target.href.endsWith('#edit')) {
        options = $graph.data('options');

        // Go to Administration Panel
        location = urlPrefix + '/admin/graphs/' + $(e.target).closest('[data-graph]').attr('data-graph');
        if (options.linked === true)
            location += '?linked=1';

        window.location = location;
    } else if (e.target.href.endsWith('#reframe-all')) {
        // Apply current options to siblings
        $graph.siblings('[data-graph]').each(function () {
            var $item = $(this),
                options = $graph.data('options');

            graphUpdateOptions($item, {
                time: options.time || null,
                range: options.range || null
            });

            graphDraw($item, !$item.inViewport());
        });

        graphDraw($graph);
    } else if (e.target.href.endsWith('#refresh')) {
        // Refresh graph
        graphDraw($graph, false);
    } else if (e.target.href.endsWith('#reset')) {
        // Reset graph timerange
        graphUpdateOptions($graph, {
            time: null,
            range: null
        });

        graphDraw($graph);
    } else if (e.target.href.endsWith('#embed')) {
        options = $graph.data('options');
        if (options.time)
            args.push('time=' + options.time.replace('+', '%2B'));
        if (options.range)
            args.push('range=' + options.range);
        if (options.refresh_interval)
            args.push('refresh=' + options.refresh_interval);

        // Open embeddable graph
        location = urlPrefix + '/show/graphs/' + $(e.target).closest('[data-graph]').attr('data-graph');
        if (args.length > 0)
            location += '?' + args.join('&');

        window.open(location);
    } else if (e.target.href.endsWith('#export')) {
        graphExport($graph);
    } else if (e.target.href.endsWith('#set-range')) {
        // Toggle range selector
        $(e.target).closest('.graphctrl').find('.ranges').toggle();
    } else if (e.target.href.endsWith('#set-time')) {
        options = $graph.data('options');

        $overlay = overlayCreate('time', {
            callbacks: {
                validate: function () {
                    graphUpdateOptions($graph, {
                        time: moment($overlay.find('input[name=time]').val()).format(TIME_RFC3339),
                        range: $overlay.find('input[name=range]').val()
                    });

                    graphDraw($graph);
                }
            }
        });

        $overlay.find('input[name=time]').appendDtpicker({
            closeOnSelected: true,
            current: options.time ? moment(options.time).format('YYYY-MM-DD HH:mm') : null,
            firstDayOfWeek: 1,
            minuteInterval: 10,
            todayButton: false
        });

        $overlay.find('input[name=range]').val(options.range || '');
    } else if (e.target.href.substr(e.target.href.lastIndexOf('#')).startsWith('#range-')) {
        range = e.target.href.substr(e.target.href.lastIndexOf('-') + 1);

        // Set graph range
        graphUpdateOptions($graph, {
            time: null,
            range: '-' + range
        });

        graphDraw($graph);
    } else if (e.target.href.endsWith('#step-backward') || e.target.href.endsWith('#step-forward')) {
        graphObj = $graph.children('.graphcntr').highcharts();

        delta = (graphObj.xAxis[0].max - graphObj.xAxis[0].min) / 4;

        if (e.target.href.endsWith('#step-backward'))
            delta *= -1;

        graphUpdateOptions($graph, {
            time: moment(graphObj.xAxis[0].min).add(delta).format(TIME_RFC3339),
            range: $graph.data('options').range.replace(/^-/, '')
        });

        graphDraw($graph);
    } else if (e.target.href.endsWith('#toggle-legend')) {
        var graphOpts = $graph.data('options') || $graph.opts('graph');

        $target.toggleClass('icon-fold icon-unfold');

        $graph.data('toggled-legend', true);

        graphUpdateOptions($graph, {
            legend: typeof graphOpts.legend == 'boolean' ? !graphOpts.legend : true
        });

        graphDraw($graph);
    } else if (e.target.href.endsWith('#zoom-in') || e.target.href.endsWith('#zoom-out')) {
        graphObj = $graph.children('.graphcntr').highcharts();

        delta = graphObj.xAxis[0].max - graphObj.xAxis[0].min;

        if (e.target.href.endsWith('#zoom-in')) {
            range = timeToRange(delta / 2);
            delta /= 4;
        } else {
            range = timeToRange(delta * 2);
            delta = (delta / 2) * -1;
        }

        graphUpdateOptions($graph, {
            time: moment(graphObj.xAxis[0].min).add(delta).format(TIME_RFC3339),
            range: range
        });

        graphDraw($graph);
    } else {
        return;
    }

    e.preventDefault();
}

function graphHandleMouse(e) {
    var $target = $(e.target),
        $graph = $target.closest('[data-graph]'),
        $control = $graph.children('.graphctrl'),
        offset;

    // Handle control lock
    if (e.type == 'mouseup' || e.type == 'mousedown') {
        GRAPH_CONTROL_LOCK = e.type == 'mousedown';
        return;
    }

    // Stop if graph has no control or is disabled
    if (GRAPH_CONTROL_LOCK || $control.length === 0 || $control.attr('disabled'))
        return;

    if (e.type != 'mousemove') {
        // Check if leaving graph
        if ($target.closest('.step, .actions').length === 0) {
            $graph.find('.graphctrl .ranges').hide();
            return;
        }

        if (GRAPH_CONTROL_TIMEOUT)
            clearTimeout(GRAPH_CONTROL_TIMEOUT);

        // Apply mask to prevent SVG events
        if (e.type == 'mouseenter')
            $control.addClass('active');
        else
            GRAPH_CONTROL_TIMEOUT = setTimeout(function () { $control.removeClass('active'); }, 1000);

        return;
    }

    // Handle steps display
    offset = $graph.offset();

    if ($target.closest('.actions').length === 0) {
        if (e.pageX - offset.left <= GRAPH_SPACING_SIZE * 2) {
            $control.find('.step a[href$=#step-backward]').addClass('active');
            return;
        } else if (e.pageX - offset.left >= $graph.width() - GRAPH_SPACING_SIZE * 2) {
            $control.find('.step a[href$=#step-forward]').addClass('active');
            return;
        }
    }

    $control.find('.step a').removeClass('active');
}

function graphHandleQueue(force) {
    var $deferreds = [];

    force = typeof force == 'boolean' ? force : false;

    if (GRAPH_DRAW_TIMEOUTS.draw)
        clearTimeout(GRAPH_DRAW_TIMEOUTS.draw);

    if (GRAPH_DRAW_TIMEOUTS.mesg)
        clearTimeout(GRAPH_DRAW_TIMEOUTS.mesg);

    return $.Deferred(function ($deferred) {
        GRAPH_DRAW_TIMEOUTS.draw = setTimeout(function () {
            var $graph,
                count = 0,
                delay = 0,
                i,
                j;

            GRAPH_DRAW_TIMEOUTS.mesg = setTimeout(function () {
                overlayCreate('loader', {
                    message: $.t('graph.mesg_loading')
                });
            }, 1000);

            for (i in GRAPH_DRAW_QUEUE) {
                for (j in GRAPH_DRAW_QUEUE[i]) {
                    if (!GRAPH_DRAW_QUEUE[i][j]) {
                        count += 1;
                        continue;
                    }

                    $graph = $(GRAPH_DRAW_QUEUE[i][j]);

                    if (force || $graph.inViewport()) {
                        $deferreds.push(graphDraw($graph, false, delay));
                        GRAPH_DRAW_QUEUE[i][j] = null;

                        if (force)
                            delay += GRAPH_DRAW_DELAY;
                    }
                }

                if (count == GRAPH_DRAW_QUEUE[i].length) {
                    GRAPH_DRAW_PARENTS.splice(i, 1);
                    GRAPH_DRAW_QUEUE.splice(i, 1);
                    $(GRAPH_DRAW_PARENTS[i]).off('scroll', graphHandleQueue);
                }
            }

            $.when.apply(null, $deferreds).then(function () {
                if (GRAPH_DRAW_TIMEOUTS.mesg)
                    clearTimeout(GRAPH_DRAW_TIMEOUTS.mesg);

                overlayDestroy('loader');
                $deferred.resolve();
            });
        }, 200);
    }).promise();
}

function graphSetupTerminate() {
    var $graphs = $('[data-graph]');

    // Get graph template
    $graphTemplate = $('.graphtmpl').removeClass('graphtmpl').detach();

    // Draw graphs
    $graphs.each(function () {
        var $item,
            id = this.getAttribute('data-graph');

        if (!id)
            return;

        $item = $(this);
        graphDraw($item, !$item.inViewport());
    });

    if ($graphs.length > 0) {
        Highcharts.setOptions({
            global : {
                useUTC : false
            }
        });
    }

    // Attach events
    $window
        .on('resize', graphHandleQueue);

    $body
        .on('mouseup mousedown mousemove mouseleave', '[data-graph]', graphHandleMouse)
        .on('mouseenter mouseleave', '.graphctrl .step, .graphctrl .actions', graphHandleMouse)
        .on('click', '[data-graph] a', graphHandleActions)
        .on('click', '.graphlist a', graphHandleQueue);
}

function graphUpdateOptions(graph, options) {
    var key;

    options = $.extend(graph.data('options'), options);

    for (key in options) {
        if (typeof options[key] != 'boolean' && !options[key])
            delete options[key];
    }

    graph.data('options', options);
}

// Register setup callbacks
setupRegister(SETUP_CALLBACK_TERM, graphSetupTerminate);
