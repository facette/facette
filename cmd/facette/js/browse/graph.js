
function browseGraphSetupTerminate() {
    // Register links
    linkRegister('edit-graph', function (e) {
        // Go to Administration Panel
        window.location = urlPrefix + '/admin/graphs/' + $(e.target).closest('[data-pane]').opts('pane').id;
    });

    linkRegister('set-global-range', browseSetRange);
    $('a[href=#set-global-range] + .menu .menuitem a').on('click', browseSetRange);
    $body.on('click', browseSetRange);

    linkRegister('set-global-refresh', browseSetRefresh);

    linkRegister('toggle-legends', browseToggleLegend);
}
