chart.fn.drawTitles = function() {
    var $$ = this;

    if (!$$.config.title && !$$.config.subtitle) {
        return;
    }

    // Create title group and texts
    $$.titleGroup = $$.mainGroup.append('g')
        .attr('transform', 'translate(' + ($$.width / 2) + ',0)');

    var title,
        subtitle;

    if ($$.config.title) {
        title = $$.titleGroup.append('text')
            .attr('class', 'chart-title')
            .attr('text-anchor', 'middle')
            .text($$.config.title);
    }

    // Stop if no subtitle defined
    if (!$$.config.subtitle) {
        return;
    }

    subtitle = $$.titleGroup.append('text')
        .attr('class', 'chart-subtitle')
        .attr('text-anchor', 'middle')
        .text($$.config.subtitle);

    if ($$.config.title) {
        subtitle.attr('transform', 'translate(0,' + title.node().getBBox().height + ')');
    }
};
