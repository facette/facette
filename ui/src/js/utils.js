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
        .replace(/[^\w\- ]+/g, '')
        .replace(/ +/g, '-');
}

function parseFloatList(input) {
    return $.map(input, function(x) { return parseFloat(x); });
}
