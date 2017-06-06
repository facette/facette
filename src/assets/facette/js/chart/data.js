chart.fn.updateData = function() {
    var $$ = this;

    // Map data set coordinates
    $$.dataSet = $$.config.series.map(function(series) {
        if (!series.plots) {
            series.disabled = true;
        }

        if (series.disabled) {
            return [];
        }

        return series.plots.map(function(a) { return {x: a[0] * 1000, y1: a[1]}; });
    });

    if ($$.config.stack && $$.dataSet.length > 0) {
        var stackData = new Array($$.dataSet[0].length),
            keys = [];

        $$.config.series.forEach(function(series) {
            if (!series.plots) {
                return;
            }

            series.plots.forEach(function(plot, idx) {
                if (!stackData[idx]) {
                    stackData[idx] = {date: plot[0] * 1000};
                }

                stackData[idx][series.name] = series.disabled ? 0 : plot[1];
            });

            keys.push(series.name);
        });

        if ($$.config.stack == 'percent') {
            // Apply percentage values
            stackData.forEach(function(entry) {
                var sum = 0;
                keys.forEach(function(key) { sum += entry[key]; });
                keys.forEach(function(key) { if (sum !== 0) { entry[key] /= sum; } });
            });
        }

        $$.stack = d3.stack()
            .keys(keys);

        if ($$.config.stack) {
            $$.stack.order(d3.stackOrderReverse);
        }

        var stackSet = new Array($$.dataSet.length).fill([]);

        $$.stack(stackData).forEach(function(entry, idx) {
            stackSet[idx] = $$.config.series[idx].disabled ?
                [] : entry.map(function(a) { return {x: a.data.date, y0: a[0], y1: a[1]}; });
        });

        $$.dataSet = stackSet;
    }
};
