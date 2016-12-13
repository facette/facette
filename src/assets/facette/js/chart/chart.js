function Chart(config, update) {
    update = typeof update == 'boolean' ? update : false;

    // Load chart configuration
    this.loadConfig(config, update);

    // Initialize data set
    this.updateData();

    // Draw chart components
    this.drawSVG();
    this.drawBackgroundRect();
    this.drawMain();
    this.drawTitles();
    this.drawAxis();
    this.drawArea();
    this.drawSeries();
    this.drawZoomRect();
    this.drawEventRect();
    this.drawTooltip();

    // Draw legend if enabled
    if (this.config.legend.enabled) {
        this.drawLegend();
    }
}

var chart = {
    fn: Chart.prototype,

    colors: [
        '#7cb5ec', '#434348', '#90ed7d', '#f7a35c', '#8085e9',
        '#f15c80', '#e4d354', '#8085e8', '#8d4653', '#91e8e1'
    ]
};

chart.create = function(config) {
    return new Chart(config, false);
};

chart.update = function(config) {
    return new Chart(config, true);
};

chart.get = function(element) {
    return element._chart;
};
