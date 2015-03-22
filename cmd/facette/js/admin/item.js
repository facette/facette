
function adminItemHandlePaneList(itemType) {
    var paneSection = paneMatch(itemType + '-list').opts('pane').section;

    // Register links
    linkRegister('show-info', function (e) {
        var $target = $(e.target),
            $item = $target.closest('[data-listitem]'),
            $tooltip;

        $tooltip = tooltipCreate('info', function (state) {
            $target.toggleClass('active', state);
            $item.toggleClass('action', state);
        }).appendTo($body)
            .css({
                top: $target.offset().top,
                left: $target.offset().left
            });

        $tooltip.html('<span class="label">id:</span> ' + $item.attr('data-itemid'));
    });

    linkRegister('show-' + itemType, function (e) {
        window.location = urlPrefix + '/browse/' + paneSection + '/' +
            $(e.target).closest('[data-itemid]').attr('data-itemid');
    });

    linkRegister('edit-' + itemType, function (e) {
        var $item = $(e.target).closest('[data-itemid]'),
            location;

        location = urlPrefix + '/admin/' + paneSection + '/' + $item.attr('data-itemid');
        if ($item.data('params'))
            location += '?' + $item.data('params');

        window.location = location;
    });

    linkRegister('clone-' + itemType, function (e) {
        var $item = $(e.target).closest('[data-itemid]');

        overlayCreate('prompt', {
            message: $.t(itemType + '.labl_' + itemType + '_name'),
            value: $item.find('.name').text() + ' (clone)',
            callbacks: {
                validate: function (data) {
                    if (!data)
                        return;

                    itemSave($item.attr('data-itemid'), paneSection, {
                        name: data
                    }, true).then(function () {
                        listUpdate(
                            $item.closest('[data-list]'),
                            $item.closest('[data-pane]').find('[data-listfilter=' + paneSection + ']').val()
                        );
                    });
                }
            },
            labels: {
                validate: {
                    text: $.t(itemType + '.labl_clone')
                }
            }
        });
    });

    linkRegister('remove-' + itemType, function (e) {
        var $item = $(e.target).closest('[data-itemid]');

        overlayCreate('confirm', {
            message: $.t(itemType + '.mesg_delete', {name: '<strong>' + $item.find('.name').text() + '</strong>'}),
            callbacks: {
                validate: function () {
                    itemDelete($item.attr('data-itemid'), paneSection)
                        .then(function () {
                            listUpdate($item.closest('[data-list]'),
                                $item.closest('[data-pane]').find('[data-listfilter=' + paneSection + ']').val());
                        })
                        .fail(function () {
                            overlayCreate('alert', {
                                message: $.t(itemType + '.mesg_delete_fail')
                            });
                        });
                }
            },
            labels: {
                validate: {
                    text: $.t(itemType + '.labl_delete'),
                    style: 'danger'
                }
            }
        });
    });
}

function adminItemHandlePaneSave(pane, itemId, itemType, callback) {
    var $item,
        paneSection = pane.opts('pane').section,
        paneParams = pane.data('redirect-params');

    $item = pane.find('input[name=' + itemType + '-name]');

    if (!$item.val()) {
        $item.closest('[data-input], textarea')
            .attr('title', $.t('main.mesg_field_mandatory'))
            .addClass('error');

        $item.focus();

        return;
    }

    itemSave(itemId, paneSection, callback(), null)
        .then(function () {
            PANE_UNLOAD_LOCK = false;
            window.location = urlPrefix + '/admin/' + paneSection + '/' + (paneParams ? '?' + paneParams : '');
        })
        .fail(function () {
            overlayCreate('alert', {
                message: $.t(itemType + '.mesg_save_fail')
            });
        });
}

function adminItemHandleReorder(e) {
    var $target = $(e.target),
        $item = $target.closest('[data-listitem]'),
        $itemNext;

    if (e.target.href.endsWith('#move-up')) {
        $itemNext = $item.prevAll('[data-listitem]:first');

        if ($itemNext.length === 0)
            return;

        $item.detach().insertBefore($itemNext);
    } else {
        $itemNext = $item.nextAll('[data-listitem]:first');

        if ($itemNext.length === 0)
            return;

        $item.detach().insertAfter($itemNext);
    }
}
