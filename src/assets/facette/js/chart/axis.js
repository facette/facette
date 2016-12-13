chart.fn.drawAxis = function() {
    var $$ = this;

    // Draw Y axis label if any
    if ($$.config.axis.y.legend) {
        $$.yLegend = $$.mainGroup.append('text')
            .attr('class', 'chart-axis-legend')
            .attr('text-anchor', 'middle')
            .attr('transform', 'translate(0,' + ($$.height / 2) + ') rotate(-90)')
            .text($$.config.axis.y.legend);
    }

    // Draw Y axis
    var yAxisTop = $$.titleGroup ? $$.titleGroup.node().getBBox().height : 0,
        yAxisLeft = $$.yLegend ? $$.yLegend.node().getBBox().height + $$.config.padding / 2 : 0;

    $$.yScale = d3.scaleLinear()
        .domain($$.getDomain('y', 'y1'))
        .range([$$.height - yAxisTop - 2 * $$.config.padding, 0])
        .nice();

    $$.yFormat = $$.config.stack == 'percent' ? d3.format('.0%') : $$.config.axis.y.tick.format;

    $$.yAxis = d3.axisLeft()
        .scale($$.yScale)
        .ticks($$.config.axis.y.tick.count)
        .tickFormat($$.yFormat);

    if ($$.yAxisGroup) {
        $$.yAxisGroup.remove();
    }

    $$.yAxisGroup = $$.mainGroup.append('g')
        .call($$.yAxis)
        .attr('class', 'chart-axis chart-axis-y')
        .attr('transform', function(a) {
            return 'translate(' + (this.getBBox().width + yAxisLeft) + ',' + yAxisTop + ')';
        });

    $$.yLines = {};

    // Draw X axis
    var xAxisTop = $$.height - 2 * $$.config.padding,
        xAxisLeft = $$.yAxisGroup.node().getBBox().width,
        xAxisWidth = $$.width - xAxisLeft - 2 * $$.config.padding;

    if ($$.yLegend) {
        var xAxisDelta = $$.yLegend.node().getBBox().height + $$.config.padding / 2;

        xAxisLeft += xAxisDelta;
        xAxisWidth -= xAxisDelta;
    }

    $$.xScale = d3.scaleTime()
        .domain($$.getDomain('x', 'x'))
        .range([0, xAxisWidth]);

    $$.xFormat = function(date) {
        return (
            d3.timeSecond(date) < date ? d3.timeFormat(".%L")
            : d3.timeMinute(date) < date ? d3.timeFormat(":%S")
            : d3.timeHour(date) < date ? d3.timeFormat("%H:%M")
            : d3.timeDay(date) < date ? d3.timeFormat("%H:00")
            : d3.timeMonth(date) < date ? d3.timeFormat("%a %d")
            : d3.timeYear(date) < date ? d3.timeFormat("%B")
            : d3.timeFormat("%Y")
        )(date);
    };

    $$.xAxis = d3.axisBottom()
        .scale($$.xScale)
        .ticks($$.config.axis.x.tick.count)
        .tickFormat($$.xFormat)
        .tickSizeOuter(0);

    if ($$.xAxisGroup) {
        $$.xAxisGroup.remove();
    }

    $$.xAxisGroup = $$.mainGroup.append('g')
        .call($$.xAxis)
        .attr('class', 'chart-axis chart-axis-x')
        .attr('transform', function(a) { return 'translate(' + xAxisLeft + ',' + xAxisTop + ')'; });
};

chart.fn.getDomain = function(axis, dataKey) {
    var $$ = this,
        min,
        max;

    if ($$.config.axis[axis].min !== null) {
        min = $$.config.axis[axis].min;
    } else {
        min = d3.min($$.dataSet, function(a) { return d3.min(a, function(b) { return b[dataKey]; }); });
        if (min > 0) {
            min = 0;
        }
    }

    if ($$.config.axis[axis].max !== null) {
        max = $$.config.axis[axis].max;
    } else {
        max = d3.max($$.dataSet, function(a) { return d3.max(a, function(b) { return b[dataKey]; }); });
    }

    // Center Y-axis zero if negative values are present
    if (axis == 'y' && min < 0) {
        max = Math.max(max, Math.abs(min));
        min = max * -1;
    }

    return [min || 0, max || 1];
};

chart.fn.addYLine = function(name, value) {
    var $$ = this;

    var xVal = $$.width - $$.yAxisGroup.node().getBBox().width - 2 * $$.config.padding,
        yVal = $$.yScale(value);

    $$.yLines[name] = $$.areaGroup.append('line')
        .attr('class', 'chart-line')
        .attr('x1', 0)
        .attr('x2', xVal)
        .attr('y1', yVal)
        .attr('y2', yVal);

    return $$.yLines[name];
};

chart.fn.removeYLine = function(name) {
    var $$ = this;

    $$.yLines[name].remove();
    delete $$.yLines[name];
};
