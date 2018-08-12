var durationUnits = ['ns', 'µs', 'ms', 's'];

function formatDuration(value, base) {
    if (value === 0) {
        return '0';
    }

    switch (base) {
    case 's':
        value *= Math.pow(10, 9);
        break;

    case 'ms':
        value *= Math.pow(10, 6);
        break;

    case 'us':
    case 'µs':
        value *= Math.pow(10, 3);
        break;
    }

    var index = parseInt(Math.log(Math.abs(value)) / Math.log(1000), 10);
    value /= Math.pow(1000, index);
    return (Math.round(value * 100) / 100) + (index > 0 ? durationUnits[index] : '');
}

var sizeUnits = ['', 'k', 'M', 'G', 'T', 'P', 'E', 'Z', 'Y'];

function formatSize(value) {
    if (value === 0) {
        return '0';
    }

    var index = parseInt(Math.log(Math.abs(value)) / Math.log(1024), 10);
    value /= Math.pow(1024, index);
    return (Math.round(value * 100) / 100) + (index > 0 ? sizeUnits[index] : '');
}

var timeUnits = {
    d: 86400000,
    h: 3600000,
    m: 60000,
    s: 1000
};

function timeToRange(duration) {
    var seconds = Math.abs(duration),
        chunks = [];

    for (var unit in timeUnits) {
        var count = Math.floor(seconds / timeUnits[unit]);
        if (count > 0) {
            chunks.push(count + unit);
            seconds %= timeUnits[unit];
        }
    }

    var result = chunks.join(' ');
    if (duration < 0) {
        result = '-' + result;
    }

    return result;
}

function slugify(input) {
    return input.toLowerCase()
        .replace(/\./g, '_')
        .replace(/[\s\-]+/g, '-')
        .replace(/[^\w\-]+/g, '');
}

function parseFloatList(input) {
    return $.map(input, function(x) { return parseFloat(x); });
}
