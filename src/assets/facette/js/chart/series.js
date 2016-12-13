chart.fn.drawSeries = function() {
    var $$ = this;

    $$.areaGroup.selectAll('.chart-series').remove();

    // Draw series paths
    var data = $$.areaGroup.selectAll('.chart-series')
        .data($$.dataSet);

    var series = data.enter()
        .insert('g', 'line.chart-cursor')
        .attr('class', 'chart-series');

    if ($$.config.type == 'area') {
        series.append('path')
            .attr('class', 'chart-area')
            .attr('d', function(a) { return $$.area(a); })
            .style('fill', function(a, i) { return chart.utils.toRGBA($$.config.series[i].color, 0.65); });
    }

    series.append('path')
        .attr('class', 'chart-line')
        .attr('d', function(a) { return $$.line(a); })
        .style('stroke', function(a, i) { return $$.config.series[i].color; });

    // Draw constants if any
    if (!$$.config.constants) {
        return;
    }

    $$.config.constants.forEach(function(constant, idx) {
        $$.addYLine('constant' + idx, constant)
            .attr('class', 'chart-line chart-constant');
    });
};

chart.fn.toggleSeries = function(idx) {
    var $$ = this;

    $$.config.series[idx].disabled = !$$.config.series[idx].disabled;

    $$.updateData();
    $$.drawAxis();
    $$.drawArea();
    $$.drawSeries();
};
