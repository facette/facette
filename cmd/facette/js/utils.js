
/* Utils */

function arrayUnique(array) {
    var result = [],
        length = array.length,
        i;

    for (i = 0; i < length; i++) {
        if (result.indexOf(array[i]) != -1)
            continue;

        result.push(array[i]);
    }

    return result;
}

function domFillItem(item, data, formatters) {
    var key;

    formatters = formatters || {};

    for (key in data) {
        if (typeof data[key] == "object")
            continue;

        item.find('.' + key).text(formatters[key] ? formatters[key](data[key]) : data[key]);
    }
}

function formatValue(value, opts) {
    var result;

    opts = opts || {};

    switch (opts.unit_type) {
    case UNIT_TYPE_FIXED:
        result = sprintf(opts.formatter || '%.2f', value);
        break;

    case UNIT_TYPE_METRIC:
        result = humanReadable(value, opts.formatter);
        break;

    default:
        result = value;
    }

    if (opts.unit)
        result += ' ' + opts.unit;

    return result;
}

function getURLParams() {
    var params = {};

    $.each(window.location.search.substr(1).split('&'), function (index, entry) {
        var pos = entry.indexOf('=');

        if (pos == -1)
            params[entry] = undefined;
        else
            params[entry.substr(0, pos)] = entry.substr(pos + 1);
    });

    return params;
}

function humanReadable(number, formatter) {
    var units = ['', 'k', 'M', 'G', 'T', 'P', 'E', 'Z', 'Y'],
        index;

    if (number === 0)
        return '0';

    index = parseInt(Math.log(Math.abs(number)) / Math.log(1000), 10);
    return sprintf(formatter || '%.2f', number / Math.pow(1000, index)) + (index > 0 ? ' ' + units[index] : '');
}

function parseFloatList(string) {
    return $.map(string.split(','), function (x) { return parseFloat(x.trim()); });
}

function rgbToHex(value) {
    var chunks;

    if (!value)
        return null;

    chunks = value.match(/^rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*\d+)?\)$/);

    return '#' +
        ('0' + parseInt(chunks[1], 10).toString(16)).slice(-2) +
        ('0' + parseInt(chunks[2], 10).toString(16)).slice(-2) +
        ('0' + parseInt(chunks[3], 10).toString(16)).slice(-2);
}

function splitAttributeValue(attrValue) {
    var entries = $.map(attrValue.split(';'), $.trim),
        i,
        index,
        key,
        result = {},
        value;

    for (i in entries) {
        index = entries[i].indexOf(':');
        key   = entries[i].substr(0, index).trim();
        value = entries[i].substr(index + 1).trim();

        if ($.isNumeric(value))
            value = parseFloat(value);
        else if (['false', 'true'].indexOf(value.toLowerCase()) != -1)
            value = value.toLowerCase() == 'true';

        result[key] = value;
    }

    return result;
}

function timeToRange(duration) {
    var units = {
            d: 86400000,
            h: 3600000,
            m: 60000,
            s: 1000
        },
        chunks = [],
        count,
        unit,
        seconds,
        result;

    seconds = Math.abs(duration);

    for (unit in units) {
        count = Math.floor(seconds / units[unit]);

        if (count > 0) {
            chunks.push(count + unit);
            seconds %= units[unit];
        }
    }

    result = chunks.join(' ');

    if (duration < 0)
        result = '-' + result;

    return result;
}

function getHighchartsSymbol(symbolText) {
    var symbol;

    switch (symbolText) {
    case 'circle':
        symbol = '●';
        break;
    case 'diamond':
        symbol = '♦';
        break;
    case 'square':
        symbol = '■';
        break;
    case 'triangle':
        symbol = '▲';
        break;
    case 'triangle-down':
        symbol = '▼';
        break;
    }
    return symbol;
}
