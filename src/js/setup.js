
/* Setup */

var SETUP_CALLBACKS     = {},
    SETUP_CALLBACK_INIT = 0,
    SETUP_CALLBACK_TERM = 1;

function setupExec(callbackType) {
    var $deferreds = [],
        i;

    for (i in SETUP_CALLBACKS[callbackType])
        $deferreds.push(SETUP_CALLBACKS[callbackType][i]());

    return $.when.apply(null, $deferreds);
}

function setupRegister(callbackType, callback) {
    if (!SETUP_CALLBACKS[callbackType])
        SETUP_CALLBACKS[callbackType] = [];

    SETUP_CALLBACKS[callbackType].push(callback);
}
