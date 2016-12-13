chart.utils = {};

chart.utils.extend = function extend(dst, src) {
    for (var key in src) {
        if (src.hasOwnProperty(key)) {
            if (dst[key] !== null && typeof dst[key] == 'object' && typeof src[key] == 'object') {
                dst[key] = chart.utils.extend(dst[key], src[key]);
            } else {
                dst[key] = src[key];
            }
        }
    }
    return dst;
};

chart.utils.toRGBA = function(color, opacity) {
    var rgb = d3.rgb(color);
    return 'rgba(' + rgb.r + ',' + rgb.g + ',' + rgb.b + ',' + (typeof opacity == 'number' ? opacity : 1) + ')';
};

chart.utils.stylesList = [
    'fill',
    'opacity',
    'stroke',
    'text-anchor'
];

chart.utils.inlineStyles = function(element) {
    var style = window.getComputedStyle(element, null),
        parts = [];

    for (var i = 0, n = style.length; i < n; i++) {
        if (chart.utils.stylesList.indexOf(style[i]) != -1) {
            parts.push(style[i] + ': ' + style.getPropertyValue(style[i]));
        }
    }

    element.style = parts.join('; ');

    for (i in element.childNodes) {
        if (element.childNodes[i].nodeType == 1) {
            chart.utils.inlineStyles(element.childNodes[i]);
        }
    }
};

chart.utils.translate = function(element) {
    if (element.transform.baseVal.length === 0) {
        return [0, 0];
    }

    var matrix = element.transform.baseVal.consolidate().matrix;

    return [matrix.e, matrix.f];
};
