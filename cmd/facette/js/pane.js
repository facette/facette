
/* Pane */

var PANE_UNLOAD_LOCK = false;

function paneGoto(pane, step, initOnly) {
    var $item,
        numeric;

    initOnly = typeof initOnly == 'boolean' ? initOnly : false;

    if (ADMIN_PANES[pane].active === null)
        ADMIN_PANES[pane].callbacks.init();

    if (!step || ADMIN_PANES[pane].callbacks['step-' + step]() === false || initOnly)
        return;

    $item = paneMatch(pane);

    $item.find('[data-step=' + step + ']')
        .show()
        .siblings('[data-step]').hide();

    numeric = !isNaN(parseInt(step, 10));

    $item.find('button[name^=step-]').toggle(numeric);
    $item.find('button[name=step-ok]').toggle(!numeric);

    if (parseInt(step, 10) == 1)
        $item.find('button[name=step-prev]').attr('disabled', 'disabled');
    else
        $item.find('button[name=step-prev]').removeAttr('disabled');

    if (parseInt(step, 10) == ADMIN_PANES[pane].count) {
        $item.find('button[name=step-next]').attr('disabled', 'disabled');
        $item.find('button[name=step-save]').removeAttr('disabled');
    } else {
        $item.find('button[name=step-next]').removeAttr('disabled');
        $item.find('button[name=step-save]').attr('disabled', 'disabled');
    }

    if (ADMIN_PANES[pane].count == 1)
        $item.find('button[name=step-prev], button[name=step-next]').hide();

    ADMIN_PANES[pane].last   = ADMIN_PANES[pane].active;
    ADMIN_PANES[pane].active = step;
}

function paneMatch(name) {
    return $('[data-pane=' + name + ']');
}

function paneRegister(pane, callback) {
    ADMIN_PANES[pane] = {
        count: 0,
        active: null,
        last: null,
        callbacks: {
            init: callback
        }
    };

    $('[data-pane=' + pane + '] [data-step]').each(function () {
        var step = parseInt(this.getAttribute('data-step'), 10);

        if (!isNaN(step) && step > ADMIN_PANES[pane].count)
            ADMIN_PANES[pane].count = step;
    });
}

function paneSetupTerminate() {
    // Initialize panes
    $('[data-pane]').each(function () {
        var name = this.getAttribute('data-pane');

        if (ADMIN_PANES[name] && ADMIN_PANES[name].active === null)
            paneGoto(name, ADMIN_PANES[name].count > 0 ? 1 : null);
    });
}

function paneStepRegister(pane, step, callback) {
    if (!ADMIN_PANES[pane]) {
        console.error("Unable to find `" + pane + "' registered pane");
        return;
    }

    ADMIN_PANES[pane].callbacks['step-' + step] = callback;
}

// Attach events
$window.on('beforeunload', function () {
    if (PANE_UNLOAD_LOCK)
        return $.t('main.mesg_unsaved_changes');
});

// Register setup callbacks
setupRegister(SETUP_CALLBACK_TERM, paneSetupTerminate);
