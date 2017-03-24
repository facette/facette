chart.fn.drawSVG = function() {
    var $$ = this,
        element = d3.select($$.config.bindTo),
        elementOld = element.selectAll('svg.chart');

    if (elementOld.node()) {
        // Preserve legend state
        $$.config.legend.enabled = chart.get(elementOld.node()).config.legend.enabled;

        // Remove existing chart
        elementOld.remove();
        element.selectAll('.chart-tooltip').remove();
    }

    // Set chart dimensions
    $$.width = $$.config.bindTo.clientWidth;
    $$.height = $$.config.bindTo.clientHeight;

    // Draw main SVG chart compoment
    $$.svg = element.append('svg')
        .attr('class', 'chart')
        .attr('width', $$.width)
        .attr('height', $$.height);

    $$.svg.node()._chart = this;
};

chart.fn.getSVG = function() {
    var node = this.svg.node().cloneNode(true);
    node.setAttribute('version', '1.1');
    node.setAttribute('xmlns', 'http://www.w3.org/2000/svg');

    // Remove UI-related nodes
    d3.select(node).selectAll('.chart-cursor, .chart-event, .chart-zoom').remove();

    chart.utils.inlineStyles(node);

    return node.outerHTML;
};
