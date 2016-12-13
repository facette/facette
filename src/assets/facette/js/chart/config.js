chart.fn.loadConfig = function(config, update) {
    update = typeof update == 'boolean' ? update : false;

    var $$ = this;

    $$.config = chart.utils.extend($$.getDefaultConfig(), config);

    if (update) {
        // Check for old state
        var element = d3.select(config.bindTo).select('svg.chart').node();
        if (element) {
            var configOld = chart.get(element).config || {},
                disableStates = {};

            configOld.series.forEach(function(series) {
                disableStates[series.name] = series.disabled || false;
            });

            $$.config.series.forEach(function(series) {
                if (disableStates[series.name]) {
                    series.disabled = disableStates[series.name];
                }
            });
        }
    }

    // Apply default series colors if none defined
    $$.config.series.forEach(function(series, idx) {
        series.color = series.color || $$.config.colors.series[idx % $$.config.colors.series.length];
    });
};

chart.fn.getDefaultConfig = function() {
    return {
        axis: {
            x: {
                max: null,
                min: null,
                tick: {
                    count: 10
                }
            },
            y: {
                legend: null,
                max: null,
                min: null,
                tick: {
                    count: 10,
                    format: d3.format('g')
                }
            }
        },
        colors: {
            series: chart.colors,
            lines: [
                '#16a085', '#27ae60', '#2980b9', '#8e44ad',
                '#2c3e50', '#f39c12', '#d35400', '#c0392b'
            ]
        },
        constants: [],
        controls: {
            enabled: true
        },
        events: {},
        legend: {
            enabled: false
        },
        padding: 24,
        series: [],
        subtitle: null,
        title: null,
        tooltip: {
            date: {
                format: null
            }
        },
        type: 'area',
        zoom: {
            enabled: false,
            onSelect: null
        }
    };
};
