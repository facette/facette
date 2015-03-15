
function browseGraphSetupTerminate() {
    // Register links
    linkRegister('edit-graph', function (e) {
        // Go to Administration Panel
        window.location = urlPrefix + '/admin/graphs/' + $(e.target).closest('[data-pane]').opts('pane').id;
    });

    linkRegister('set-refresh', browseSetRefresh);
}
