
/* Stats */

function statsInit(element) {
    return $.Deferred(function ($deferred) {
        // Initialize stats content
        statsUpdate($(element)).then(function () { $deferred.resolve(); });
    }).promise();
}

function statsSetupInit() {
    return $.Deferred(function ($deferred) {
        var $deferreds = [];

        $('[data-stats]').each(function () {
            $deferreds.push(statsInit(this));
        });

        $.when.apply(null, $deferreds).then(function () { $deferred.resolve(); });
    }).promise();
}

function statsUpdate(stats) {
    return $.ajax({
        url: '/stats',
        type: 'GET'
    }).pipe(function (data) {
        domFillItem(stats, data, {
            catalog_updated: function (x) { return moment(x).format('LLL'); }
        });
    });
}

// Register setup callbacks
setupRegister(SETUP_CALLBACK_INIT, statsSetupInit);
