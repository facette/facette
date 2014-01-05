
/* Utils */

function domFillItem(item, data, formatters) {
    var key;

    formatters = formatters || {};

    for (key in data)
        item.find('.' + key).text(formatters[key] ? formatters[key](data[key]) : data[key]);
}

function humanReadable(number) {
    var units = ['', 'k', 'M', 'G', 'T', 'P', 'E', 'Z', 'Y'],
        index = parseInt(Math.log(number) / Math.log(1000), 10);

    return (Math.round((number / Math.pow(1000, index) * 100)) / 100) + units[index];
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

function splitAttributeValue(value) {
    var entries = $.map(value.split(';'), $.trim),
        chunks,
        i,
        result = {};

    for (i in entries) {
        chunks = $.map(entries[i].split(':'), $.trim);
        result[chunks[0]] = chunks[1];
    }

    return result;
}

function timeToRange(duration) {
    var ranges = {
            d: 86400000,
            h: 3600000,
            m: 60000,
            s: 1000
        },
        chunks = [],
        count,
        unit;

    for (unit in ranges) {
        count = Math.floor(duration / ranges[unit]);

        if (count > 0) {
            chunks.push(count + unit);
            duration %= ranges[unit];
        }
    }

    return chunks.join(' ');
}
