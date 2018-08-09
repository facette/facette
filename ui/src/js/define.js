var filterActions = [
        'discard',
        'rewrite',
        'sieve'
    ],

    filterTargets = [
        'all',
        'metric',
        'origin',
        'source'
    ],

    stateOK = 1,
    stateLoading = 2,
    stateError = 3,

    dialogTypeConfirm = 'confirm',
    dialogTypePrompt = 'prompt',

    groupPatternSingle = 1,
    groupPatternGlob = 2,
    groupPatternRegexp = 3,

    groupOperatorNone = 0,
    groupOperatorAverage = 1,
    groupOperatorSum = 2,

    groupConsolidateAverage = 1,
    groupConsolidateFirst = 2,
    groupConsolidateLast = 3,
    groupConsolidateMax = 4,
    groupConsolidateMin = 5,
    groupConsolidateSum = 6,

    patternPrefixGlob = 'glob:',
    patternPrefixRegexp = 'regexp:',

    pagingLimit = 20,

    graphMargin = 24,

    graphTypeArea = 'area',
    graphTypeLine = 'line',

    graphYAxisUnitFixed = 'fixed',
    graphYAxisUnitMetric = 'metric',
    graphYAxisUnitBinary = 'binary',

    graphStackModeNone = null,
    graphStackModeNormal = 'normal',
    graphStackModePercent = 'percent',

    timeRanges = [
        '1h',
        '3h',
        '1d',
        '3d',
        '7d',
        '1mo',
        '3mo',
        '1y'
    ],

    graphSummaryBase = [
        'min',
        'avg',
        'max',
        'last'
    ],

    groupPrefix = 'group:',

    timeFormatDisplay = 'MMMM D YYYY, HH:mm:ss',
    timeFormatFilename = 'YYYYMMDDHHmmss',
    timeFormatRFC3339 = 'YYYY-MM-DDTHH:mm:ss.SSSZ',

    templateRegexp = /\{\{\s*\.([a-z0-9]+)\s*\}\}/i,

    sidebarCollapseWith = 768;
