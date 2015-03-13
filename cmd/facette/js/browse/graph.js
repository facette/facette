
function browseGraphSetupTerminate() {
    // Register links
    linkRegister('edit-graph', function (e) {
        var opts = $(e.target).closest('[data-pane]').opts('pane'),
            location;

        // Go to Administration Panel
        location = urlPrefix + '/admin/graphs/' + opts.id;
        if (opts.linked === true)
            location += '?linked=1';

        window.location = location;
    });

    linkRegister('set-global-range', browseSetRange);
    $('a[href=#set-global-range] + .menu .menuitem a').on('click', browseSetRange);
    $body.on('click', browseSetRange);

    linkRegister('set-global-refresh', browseSetRefresh);

    linkRegister('toggle-legends', browseToggleLegend);
}
