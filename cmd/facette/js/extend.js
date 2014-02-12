
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
        var chart = this,
            options = chart.options,
            cellLeft,
            columnKeys = ['min', 'avg', 'max', 'last'],
            groups = {},
            tableLeft = chart.plotLeft,
            tableTop = chart.plotTop + chart.plotHeight + options.chart.spacingBottom -
                chart.series.length * GRAPH_LEGEND_ROW_HEIGHT;

        cellLeft = tableLeft;

        // Render custom legend
        $.each(chart.series, function (i, serie) {
            var box,
                element,
                keys;

            groups[serie.name] = chart.renderer.g().add();

            chart.renderer.rect(tableLeft, tableTop + i * GRAPH_LEGEND_ROW_HEIGHT, GRAPH_LEGEND_ROW_HEIGHT * 0.75,
                    GRAPH_LEGEND_ROW_HEIGHT * 0.65, 2)
                .attr({
                    fill: serie.color
                })
                .add(groups[serie.name]);

            element = chart.renderer.text(serie.name, tableLeft + GRAPH_LEGEND_ROW_HEIGHT, tableTop +
                    i * GRAPH_LEGEND_ROW_HEIGHT + GRAPH_LEGEND_ROW_HEIGHT / 2)
                .attr({
                    'class': 'highcharts-table-serie'
                })
                .css({
                    cursor: 'pointer'
                })
                .add(groups[serie.name])
                .element;

            Highcharts.addEvent(element, 'click', function () {
                var $element = $(element),
                    serie = chart.get($element.text());

                serie.setVisible(!serie.visible);
                $element.closest('g').css('opacity', serie.visible ? 1 : 0.35);
            });

            // Update start position
            box = element.getBBox();
            cellLeft = Math.max(cellLeft, box.x + box.width + GRAPH_LEGEND_ROW_HEIGHT);

            // Update column keys list
            if (!data[serie.name])
                return;

            keys = Object.keys(data[serie.name]);
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

            $.each(chart.series, function (i, serie) {
                var value;

                element = chart.renderer.text(key, keyLeft, tableTop + i * GRAPH_LEGEND_ROW_HEIGHT +
                        GRAPH_LEGEND_ROW_HEIGHT / 2)
                    .attr({
                        'class': 'highcharts-table-label'
                    })
                    .css({
                        color: options.plotOptions.area.dataLabels.style.color
                    })
                    .add(groups[serie.name])
                    .element;

                if (valueLeft === 0) {
                    box = element.getBBox();
                    valueLeft = box.x + box.width + GRAPH_LEGEND_ROW_HEIGHT * 0.35;
                }

                value = data[serie.name] && data[serie.name][key] ? data[serie.name][key] : null;

                element = chart.renderer.text(value !== null ? humanReadable(value) : 'null', valueLeft,
                        tableTop + i * GRAPH_LEGEND_ROW_HEIGHT + GRAPH_LEGEND_ROW_HEIGHT / 2)
                    .attr({
                        'class': 'highcharts-table-value',
                        'data-value': value
                    })
                    .css({
                        cursor: 'pointer'
                    })
                    .add(groups[serie.name])
                    .element;

                Highcharts.addEvent(element, 'click', function (e) {
                    if (options.chart.events && options.chart.events.togglePlotLine)
                        options.chart.events.togglePlotLine.apply({
                            chart: chart,
                            element: e.target,
                            name: key,
                            serie: serie,
                            value: parseFloat($(e.target).parent().attr('data-value')) || null
                        });
                });

                box = element.getBBox();
                cellLeft = Math.max(cellLeft, box.x + box.width + GRAPH_LEGEND_ROW_HEIGHT);
            });
        });
    };
}
