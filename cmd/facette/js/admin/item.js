
function adminItemHandlePaneList(itemType) {
    var paneSection = paneMatch(itemType + '-list').opts('pane').section;

    // Register links
    linkRegister('show-' + itemType, function (e) {
        window.location = urlPrefix + '/browse/' + paneSection + '/' +
            $(e.target).closest('[data-itemid]').attr('data-itemid');
    });

    linkRegister('edit-' + itemType, function (e) {
        window.location = urlPrefix + '/admin/' + paneSection + '/' +
            $(e.target).closest('[data-itemid]').attr('data-itemid');
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
                        listUpdate($item.closest('[data-list]'),
                            $item.closest('[data-pane]')
                                .find('[data-listfilter=' + paneSection + ']').val());
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
                            listUpdate($item.closest('[data-list]'));
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
    var paneSection = paneMatch(itemType + '-edit').opts('pane').section,
        skip = false;

    pane.find('input[name=' + itemType + '-name]').each(function () {
        var $item = $(this);

        if (!$item.val()) {
            $item.closest('[data-input], textarea')
                .attr('title', $.t('main.mesg_field_mandatory'))
                .addClass('error');

            skip = true;
        }
    });

    if (skip) {
        return;
    }

    itemSave(itemId, paneSection, callback(), null)
        .then(function () {
            PANE_UNLOAD_LOCK = false;
            window.location = urlPrefix + '/admin/' + paneSection + '/';
        })
        .fail(function () {
            overlayCreate('alert', {
                message: $.t(itemType + '.mesg_save_fail')
            });
        });
}
