chart.fn.drawMain = function() {
    var $$ = this;

    $$.mainGroup = $$.svg.append('g')
        .attr('transform', 'translate(' + $$.config.padding  + ',' + $$.config.padding + ')');
};
