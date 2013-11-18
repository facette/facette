
function browseCollectionSetupTerminate() {
    $('[data-treeitem=' + paneMatch('collection-show').opts('pane').id + ']')
        .addClass('current')
        .parentsUntil('[data-tree]').show();

    // Attach events
    $body
        .on('click', '[data-tree=collections] a', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-treeitem]');

            if ($target.closest('.icon').length === 0 || !$item.hasClass('folded') && !$item.hasClass('unfolded')) {
                window.location = '/browse/collections/' + $item.attr('data-treeitem');
            } else {
                $item.children('.treecntr')
                    .toggle()
                    .toggleClass('folded unfolded');
            }

            e.preventDefault();
            e.stopPropagation();
        });
}
