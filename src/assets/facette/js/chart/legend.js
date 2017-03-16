chart.fn.drawLegend = function() {
    var $$ = this;

    if ($$.config.series.length === 0) {
        return;
    }

    var areaBBox = $$.areaGroup.node().getBBox(),
        legendTop = ($$.titleGroup ? $$.titleGroup.node().getBBox().height : 0) +
            areaBBox.height + $$.xAxisGroup.node().getBBox().height + $$.config.padding / 2,
        legendLeft = $$.yAxisGroup.node().getBBox().width,
        legendLineHeight = 24,
        legendBoundary = legendLeft + areaBBox.width;

    $$.legendGroup = $$.mainGroup.append('g')
        .attr('class', 'chart-legend')
        .attr('transform', 'translate(' + legendLeft + ',' + legendTop + ')');

    // Loop through series
    function toggleLegendSeries(idx) {
        var parent = d3.select(this.parentNode);

        // Toggle series display
        if (d3.event.shiftKey) {
            $$.selectSeries(idx);

            d3.select(this.parentNode.parentNode).selectAll('.chart-legend-row').classed('disabled', true);
            parent.classed('disabled', false);
        } else {
            $$.toggleSeries(idx);

            parent.classed('disabled', !parent.classed('disabled'));
        }
    }

    function toggleLegendYLine(data) {
        // Stop if series disabled or value is null
        if (d3.select(this.parentNode).classed('disabled') ||
                $$.config.series[data[0]].summary[data[1]] === null) {
            return;
        }

        var name = data[1] + '\x1e' + data[2],
            node = d3.select(this);

        // Remove line if already present
        if (node.classed('active')) {
            $$.removeYLine(name);
            node.attr('fill', 'inherit').classed('active', false);
            return;
        }

        var line = $$.addYLine(name, data[2]),
            color = $$.config.colors.lines[Object.keys($$.yLines).indexOf(name) %
                $$.config.colors.lines.length];

        line.attr('stroke', color);
        node.attr('fill', color).classed('active', true);
    }

    var legendRows = [],
        legendColumns = [],
        columnLeft = 0,
        i,
        j,
        element,
        elementBBox;

    var filterKeys = function(a) {
        return graphSummaryBase.indexOf(a) == -1;
    };

    var series = $$.config.series;
    if ($$.config.stack) {
        series.reverse();
    }

    series.forEach(function(entry, idx) {
        legendRows[idx] = $$.legendGroup.append('g')
            .attr('class', 'chart-legend-row')
            .attr('transform', 'translate(0,' + (idx * legendLineHeight) + ')')
            .classed('disabled', entry.disabled || false);

        legendRows[idx].append('rect')
            .attr('class', 'chart-legend-color')
            .attr('width', legendLineHeight * 0.65)
            .attr('height', legendLineHeight * 0.5)
            .attr('rx', 2)
            .attr('ry', 2)
            .attr('y', legendLineHeight * 0.25)
            .attr('fill', series[idx].color);

        element = legendRows[idx].append('text')
            .datum(idx)
            .attr('class', 'chart-legend-name')
            .attr('x', legendLineHeight)
            .attr('y', legendLineHeight / 2)
            .text(entry.name)
            .on('click', toggleLegendSeries);

        // Update column left position
        elementBBox = element.node().getBBox();
        columnLeft = Math.max(columnLeft, elementBBox.x + elementBBox.width + legendLineHeight);

        // Stop if no summary data
        if (!entry.summary) {
            return;
        }

        // Retrieve legend keys
        var keys = Object.keys(entry.summary);
        keys.sort();
        keys = graphSummaryBase.concat(keys.filter(filterKeys));

        keys.forEach(function(key) {
            if (legendColumns.indexOf(key) == -1) {
                legendColumns.push(key);
            }
        });
    });

    var rowDelta = 0;

    legendColumns.forEach(function(key) {
        var groupBBox;

        var groupTop = rowDelta * legendLineHeight + legendLineHeight / 2,
            keyLeft = columnLeft,
            valueLeft = 0;

        $$.config.series.forEach(function(series, idx) {
            var group = legendRows[idx].append('g')
                .attr('class', 'chart-legend-group')
                .attr('transform', 'translate(' + keyLeft + ',' + groupTop + ')');

            // Draw summary key
            element = group.append('text')
                .attr('class', 'chart-legend-key')
                .text(key);

            // Update value left position if first one
            if (valueLeft === 0) {
                valueLeft = group.node().getBBox().width + legendLineHeight * 0.35;
            }

            // Draw summary value
            var value = series.summary && series.summary[key] !== undefined ? series.summary[key] : null;

            element = group.append('text')
                .datum([idx, key, value])
                .attr('class', 'chart-legend-value')
                .attr('x', valueLeft)
                .text(value === null ? 'null' : $$.config.axis.y.tick.format(value));

            if (value !== null) {
                element.on('click', toggleLegendYLine);
            }

            // Update column left position
            groupBBox = group.node().getBBox();
            columnLeft = Math.max(columnLeft, keyLeft + groupBBox.x + groupBBox.width + legendLineHeight * 0.65);

            if (columnLeft > legendBoundary) {
                rowDelta += 1;

                groupTop = rowDelta * legendLineHeight + legendLineHeight / 2;
                columnLeft = keyLeft = 0;

                group.attr('transform', 'translate(0,' + groupTop + ')');
            }
        });
    });

    // Handle legend rows delta
    if (rowDelta > 0) {
        legendRows.forEach(function(row, idx) {
            if (idx === 0) {
                return;
            }

            var translate = chart.utils.translate(row.node());

            row.attr('transform', 'translate(' + translate[0] + ',' +
                (translate[1] + idx * rowDelta * legendLineHeight) + ')');
        });
    }

    // Update chart height
    var newHeight = $$.height + $$.legendGroup.node().getBBox().height + $$.config.padding;
    $$.svg.attr('height', newHeight);
    $$.bgRect.attr('height', newHeight);
};

chart.fn.removeLegend = function() {
    var $$ = this;

    // Remove legend group
    $$.legendGroup.remove();
    delete $$.legendGroup;

    // Reset chart height
    $$.svg.attr('height', $$.height);
    $$.bgRect.attr('height', $$.height);
};

chart.fn.toggleLegend = function(state) {
    var $$ = this;

    // Update legend configuration
    $$.config.legend.enabled = state;

    // Toggle legend display
    if (state && !$$.legendGroup) {
        $$.drawLegend();
    } else if (!state && $$.legendGroup) {
        $$.removeLegend();
    }
};
