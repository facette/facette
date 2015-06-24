
/* Extend: base */

if (!String.prototype.trim) {
    String.prototype.trim = function () {
        return this.replace(/^\s+|\s+$/g, '');
    };
}

if (!String.prototype.startsWith) {
    String.prototype.startsWith = function (string) {
        return this.substr(0, string.length) === string;
    };
}

if (!String.prototype.endsWith) {
    String.prototype.endsWith = function (string) {
        return this.substr(-(string.length)) === string;
    };
}

if (!String.prototype.matchAll) {
    String.prototype.matchAll = function (regexp) {
        var matches = [],
            match,
            regexpStr,
            index;

        // Force RegExp global flag
        if (!regexp.global) {
            regexpStr = regexp.toString();
            index = regexpStr.lastIndexOf('/');
            regexp = new RegExp(regexpStr.substr(regexpStr.indexOf('/') + 1, index - 1), regexpStr.substr(index + 1) + 'g');
        }

        while ((match = regexp.exec(this))) {
            if (match.length === 0)
                continue;

            matches.push(match[1]);
        }

        return matches;
    };
}

/* Extend: jQuery */

$.event.props.push('dataTransfer');

$.fn.extend({
    inViewport: function () {
        var $element = $(this),
            viewTop = $window.scrollTop(),
            viewBottom = viewTop + $window.height(),
            elementTop = $element.offset().top,
            elementBottom = elementTop + $element.height();

        return elementTop <= viewBottom && elementBottom >= viewTop;
    },

    opts: function (attributeName) {
        if (!this.attr('data-' + attributeName + 'opts'))
            return {};

        return splitAttributeValue(this.attr('data-' + attributeName + 'opts'));
    }
});

/* Highcharts */
if (window.Highcharts) {
    Highcharts.drawTable = function (data) {
        var $container,
            chart = this,
            options = chart.options,
            cellLeft,
            columnKeys = ['min', 'avg', 'max', 'last'],
            groups = {},
            tableLeft = chart.plotLeft,
            tableTop = chart.plotTop + chart.plotHeight + GRAPH_SPACING_SIZE * 2.5,
            groupTimeout = {},
            groupEvent = function (e) {
                var $group = $(e.target).closest('.highcharts-table-group'),
                    series = $group.find('.highcharts-table-series').text();

                if (groupTimeout[series])
                    clearTimeout(groupTimeout[series]);

                groupTimeout[series] = setTimeout(function () {
                    if (e.type == 'mouseover')
                        $group.parent().find('.highcharts-table-action').css('visibility', 'hidden');

                    $group.children('.highcharts-table-action')
                        .css('visibility', e.type == 'mouseover' ? 'visible' : 'hidden');
                }, e.type == 'mouseenter' ? 0 : 500);
            };

        cellLeft = tableLeft;

        $container = $(chart.container);

        // Clean up previous table
        $container.find('.highcharts-table-group').remove();

        // Render custom legend
        $.each(chart.series, function (i, series) {
            var box,
                element,
                keys;

            groups[series.name] = chart.renderer.g()
                .attr({
                    'class': 'highcharts-table-group'
                })
                .add();

            if (!series.visible)
                $(groups[series.name].element).css('opacity', 0.35);

            element = chart.renderer.text('\uf176', tableLeft - GRAPH_LEGEND_ROW_HEIGHT * 0.5, tableTop +
                    i * GRAPH_LEGEND_ROW_HEIGHT + GRAPH_LEGEND_ROW_HEIGHT / 2)
                .attr({
                    'class': 'highcharts-table-action',
                    color: options.plotOptions.area.dataLabels.style.color
                })
                .css({
                    cursor: 'pointer',
                    display: 'none',
                    fontFamily: 'FontAwesome',
                    opacity: 0.25,
                    visibility: 'hidden'
                })
                .add(groups[series.name])
                .element;

            Highcharts.addEvent(element, 'click', function () {
                var $element = $(element),
                    series = chart.get($element.text());

                series.group.toFront();
            });

            Highcharts.addEvent(element, 'mouseenter mouseout', function (e) {
                $(e.target).css('opacity', e.type == 'mouseenter' ? 1 : 0.25);
                groupEvent(e);
            });

            element = chart.renderer.text(getHighchartsSymbol(series.symbol), tableLeft, tableTop +
                    i * GRAPH_LEGEND_ROW_HEIGHT + GRAPH_LEGEND_ROW_HEIGHT / 2)
                .css({
                    'color': series.color,
                })
                .add(groups[series.name])
                .element;

            Highcharts.addEvent(element, 'mouseenter mouseout', groupEvent);

            element = chart.renderer.text(series.name, tableLeft + GRAPH_LEGEND_ROW_HEIGHT, tableTop +
                    i * GRAPH_LEGEND_ROW_HEIGHT + GRAPH_LEGEND_ROW_HEIGHT / 2)
                .attr({
                    'class': 'highcharts-table-series'
                })
                .css({
                    cursor: 'pointer'
                })
                .add(groups[series.name])
                .element;

            Highcharts.addEvent(element, 'click', function () {
                var series = chart.get($(element).text());
                series.setVisible(!series.visible);
            });

            Highcharts.addEvent(element, 'mouseenter mouseout', groupEvent);

            // Update start position
            box = element.getBBox();
            cellLeft = Math.max(cellLeft, box.x + box.width + GRAPH_LEGEND_ROW_HEIGHT);

            // Update column keys list
            if (!data[series.name] || !data[series.name].summary)
                return;

            keys = Object.keys(data[series.name].summary);
            keys.sort();

            $.each(keys, function (i, key) { /*jshint unused: true */
                if (columnKeys.indexOf(key) != -1)
                    return;

                columnKeys.push(key);
            });
        });

        $.each(columnKeys, function (i, key) { /*jshint unused: true */
            var box,
                element,
                keyLeft = cellLeft,
                valueLeft = 0;

            $.each(chart.series, function (i, series) {
                var value,
                    elementOpts;

                // Draw series summary item label
                element = chart.renderer.text(key, keyLeft, tableTop + i * GRAPH_LEGEND_ROW_HEIGHT +
                        GRAPH_LEGEND_ROW_HEIGHT / 2)
                    .attr({
                        'class': 'highcharts-table-label'
                    })
                    .css({
                        color: options.plotOptions.area.dataLabels.style.color
                    })
                    .add(groups[series.name])
                    .element;

                if (valueLeft === 0) {
                    box = element.getBBox();
                    valueLeft = box.x + box.width + GRAPH_LEGEND_ROW_HEIGHT * 0.35;
                }

                // Get summary item value
                value = data[series.name] && data[series.name].summary && data[series.name].summary[key] !== undefined ?
                    data[series.name].summary[key] : null;

                // Generate element options object
                elementOpts = {unit_type: options._opts.unit_type};
                if (data[series.name].options)
                    elementOpts = $.extend(elementOpts, data[series.name].options);

                // Render summary item value
                element = chart.renderer.text(value !== null ? formatValue(value, elementOpts) : 'null',
                        valueLeft, tableTop + i * GRAPH_LEGEND_ROW_HEIGHT + GRAPH_LEGEND_ROW_HEIGHT / 2)
                    .attr({
                        'class': 'highcharts-table-value',
                        'data-name': series.name + '-' + key,
                        'data-value': value
                    })
                    .css({
                        cursor: 'pointer'
                    })
                    .add(groups[series.name])
                    .element;

                // Add item event
                Highcharts.addEvent(element, 'click', function (e) {
                    if (options.chart.events && options.chart.events.togglePlotLine)
                        options.chart.events.togglePlotLine.apply({
                            chart: chart,
                            element: e.target,
                            name: key,
                            series: series,
                            value: parseFloat($(e.target).closest('.highcharts-table-value').attr('data-value')) || null
                        });
                });

                // Calculate future position boundaries
                box = element.getBBox();
                cellLeft = Math.max(cellLeft, box.x + box.width + GRAPH_LEGEND_ROW_HEIGHT);
            });
        });

        // Attach events
        $container.closest('[data-graph]').on('mouseenter mouseleave', function (e) {
            $container.find('.highcharts-table-action').toggle(e.type == 'mouseenter');
        });
    };
}
