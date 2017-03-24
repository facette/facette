chart.fn.drawBackgroundRect = function() {
    var $$ = this;

    $$.bgRect = $$.svg.append('rect')
        .attr('class', 'chart-background')
        .attr('width', $$.width)
        .attr('height', $$.height);
};

chart.fn.drawZoomRect = function() {
    var $$ = this;

    if (!$$.config.zoom.enabled) {
        return;
    }

    // Initialize zoom rectangle
    $$.zoomRect = $$.areaGroup.append('rect')
        .attr('class', 'chart-zoom')
        .attr('width', 0)
        .attr('height', $$.height - ($$.titleGroup ? $$.titleGroup.node().getBBox().height : 0) - 2 * $$.config.padding)
        .style('display', 'none');

    $$.zoomOrigin = null;
    $$.zoomActive = false;
    $$.zoomCancelled = false;
};

chart.fn.resetZoomRect = function() {
    var $$ = this;

    // Detach key event
    d3.select(document.body).on('keydown', null);

    // Reset zoom selection
    $$.zoomOrigin = null;
    $$.zoomActive = false;
    $$.zoomCancelled = false;

    $$.zoomRect
        .attr('x', 0)
        .attr('width', 0)
        .style('display', 'none');

    // Restore cursor line
    $$.cursorLine.style('display', 'block');
};

chart.fn.drawEventRect = function() {
    var $$ = this;

    var dateBisect = d3.bisector(function(a) { return a[0] * 1000; }).left,
        rectWidth = $$.areaGroup.node().getBBox().width,
        rectHeight = $$.height - ($$.titleGroup ? $$.titleGroup.node().getBBox().height : 0) - 2 * $$.config.padding;

    $$.areaGroup.append('rect')
        .attr('class', 'chart-event')
        .attr('fill', 'transparent')
        .attr('width', rectWidth)
        .attr('height', rectHeight)
        .on('dragstart', function() {
            d3.event.preventDefault();
            d3.event.stopPropagation();
        })
        .on('mouseout', function() {
            // Hide tooltip and cursor line
            $$.cursorLine.style('display', 'none');
            $$.toggleTooltip(false);

            if ($$.config.events.cursorMove) {
                $$.config.events.cursorMove(null);
            }
        })
        .on('mousedown', function() {
            if ($$.config.zoom.onStart) {
                $$.config.zoom.onStart();
            }

            // Attach key event
            d3.select(document.body).on('keydown', function() {
                if (d3.event.which == 27) { // <Escape>
                    $$.zoomCancelled = true;
                    $$.resetZoomRect();
                }
            });

            // Initialize zoom selection
            $$.zoomOrigin = d3.mouse(this)[0];
            $$.zoomActive = true;
            $$.zoomRect.style('display', 'block');

            // Hide cursor line during selection
            $$.cursorLine.style('display', 'none');
        })
        .on('mouseup', function() {
            // Execute callback
            if (!$$.zoomCancelled && $$.config.zoom.onSelect) {
                var start = $$.xScale.invert(parseInt($$.zoomRect.attr('x'), 10)),
                    end = $$.xScale.invert(parseInt($$.zoomRect.attr('x'), 10) +
                        parseInt($$.zoomRect.attr('width'), 10));

                var startTime = start.getTime(),
                    endTime = end.getTime();

                if (!isNaN(startTime) && !isNaN(endTime) && startTime !== endTime) {
                    $$.config.zoom.onSelect(start, end);
                }
            }

            $$.resetZoomRect();
        })
        .on('mousemove', function() {
            var mouse = d3.mouse(this),
                tooltipPos = mouse[0] + $$.yAxisGroup.node().getBBox().width + $$.config.padding,
                tooltipPosKey = 'left',
                tooltipWidth = $$.tooltipGroup.node().clientWidth;

            // Update zoom selection if active
            if ($$.zoomActive) {
                $$.zoomRect
                    .attr('x', Math.min($$.zoomOrigin, mouse[0]))
                    .attr('width', Math.abs(mouse[0] - $$.zoomOrigin));
            }

            // Show tooltip and cursor line
            if (!$$.tooltipEnabled) {
                $$.cursorLine.style('display', 'block');
                $$.toggleTooltip(true);
            }

            // Set cursor line position
            $$.cursorLine
                .attr('x1', mouse[0])
                .attr('x2', mouse[0]);

            if ($$.config.events.cursorMove) {
                $$.config.events.cursorMove($$.xScale.invert(mouse[0]));
            }

            // Update tooltip position
            if (tooltipPos + tooltipWidth > rectWidth) {
                tooltipPos = Math.abs(mouse[0] - rectWidth) + $$.config.padding;
                tooltipPosKey = 'right';

                $$.tooltipGroup.style('left', null);
            } else {
                $$.tooltipGroup.style('right', null);
            }

            $$.tooltipGroup
                .style(tooltipPosKey, tooltipPos + 'px')
                .style('top', mouse[1] + 'px');

            // Set tooltip content
            var data = {
                date: $$.xScale.invert(mouse[0]),
                values: []
            };

            $$.config.series.forEach(function(series, idx) {
                var idxPlot = series.plots ? dateBisect(series.plots, data.date, 1) : -1;

                data.values[idx] = {
                    name: series.name,
                    value: idxPlot != -1 ? series.plots[idxPlot] : null
                };
            });

            $$.updateTooltip(data);
        });
};
