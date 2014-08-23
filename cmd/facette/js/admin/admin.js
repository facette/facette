
function adminHandleFieldType(e) {
    if (ADMIN_FIELD_TIMEOUT)
        clearTimeout(ADMIN_FIELD_TIMEOUT);

    ADMIN_FIELD_TIMEOUT = setTimeout(function () {
        $(e.target).trigger({
            type: 'change',
            _typing: true
        });
    }, 200);
}

function adminHandlePaneStep(e, name) {
    name = $(e.target).closest('[data-pane]').attr('data-pane');

    if (e.target.name == 'step-ok')
        paneGoto(name, ADMIN_PANES[name].last);
    else if (e.target.name == 'step-prev' && ADMIN_PANES[name].active > 1)
        paneGoto(name, ADMIN_PANES[name].active - 1);
    else if (e.target.name == 'step-next' && ADMIN_PANES[name].active < ADMIN_PANES[name].count)
        paneGoto(name, ADMIN_PANES[name].active + 1);
}
