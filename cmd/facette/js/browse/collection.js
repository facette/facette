
function browseCollectionSetupTerminate() {
    $('[data-treeitem=' + paneMatch('collection-show').opts('pane').id + ']')
        .addClass('current')
        .parentsUntil('[data-tree]').show();

    // Register links
    linkRegister('edit-collection', function (e) {
        // Go to Administration Panel
        window.location = urlPrefix + '/admin/collections/' + $(e.target).closest('[data-pane]').opts('pane').id;
    });

    linkRegister('set-refresh', browseSetRefresh);

    // Attach events
    $body
        .on('click', '[data-tree=collections] a', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-treeitem]');

            if ($target.closest('.icon').length === 0 || !$item.hasClass('folded') && !$item.hasClass('unfolded')) {
                window.location = urlPrefix + '/browse/collections/' + $item.attr('data-treeitem');
            } else {
                $item.children('.treecntr')
                    .toggle()
                    .toggleClass('folded unfolded');
            }

            e.preventDefault();
            e.stopPropagation();
        });
}
