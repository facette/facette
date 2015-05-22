
function browseCollectionHandleTree(e) {
    var $target = $(e.target),
        $item = $target.closest('[data-treeitem]');

    if ($target.closest('.icon').length === 0 || !$item.hasClass('folded') && !$item.hasClass('unfolded')) {
        window.location = urlPrefix + '/browse/collections/' + $item.attr('data-treeitem');
    } else {
        $item.toggleClass('folded unfolded')
            .children('.treecntr').toggle();
    }

    e.preventDefault();
    e.stopPropagation();
}

function browseCollectionSetupTerminateTree() {
    // Attach events
    $body.on('click', '[data-tree=collections] a', browseCollectionHandleTree);
}

function browseCollectionSetupTerminate() {
    $('[data-treeitem=' + paneMatch('collection-show').opts('pane').id + ']')
        .addClass('current')
        .parentsUntil('[data-tree]').show();

    // Register links
    linkRegister('edit-collection', function (e) {
        // Go to Administration Panel
        window.location = urlPrefix + '/admin/collections/' + $(e.target).closest('[data-pane]').opts('pane').id;
    });

    linkRegister('set-global-range', browseSetRange);
    $('a[href=#set-global-range] + .menu .menuitem a').on('click', browseSetRange);
    $body.on('click', browseSetRange);

    linkRegister('set-global-refresh', browseSetRefresh);

    linkRegister('toggle-legends', browseToggleLegend);
}
