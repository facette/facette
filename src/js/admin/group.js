
function adminGroupCreateItem(value) {
    var $item = listAppend(listMatch('step-1-items'))
        .attr('data-item', value.id)
        .data('value', value);

    domFillItem($item, value);

    return $item;
}

function adminGroupGetData() {
    var $pane = paneMatch('group-edit'),
        data = {
            name: $pane.find('input[name=group-name]').val(),
            description: $pane.find('textarea[name=group-desc]').val(),
            entries: []
        };

    listMatch('step-1-items').find('[data-listitem^=step-1-items-item]').each(function () {
        data.entries.push($(this).data('value'));
    });

    return data;
}

function adminGroupSetupTerminate() {
    var completionCallbacks;

    // Register admin panes
    paneRegister('group-list', function () {
        var groupType = paneMatch('group-list').opts('pane').section;

        // Register links
        linkRegister('edit-group', function (e) {
            window.location = '/admin/' + groupType + '/' + $(e.target).closest('[data-itemid]').attr('data-itemid');
        });

        linkRegister('clone-group', function (e) {
            var $item = $(e.target).closest('[data-itemid]');

            overlayCreate('prompt', {
                message: $.t('group.labl_group_name'),
                value: $item.find('.name').text() + ' (clone)',
                callbacks: {
                    validate: function (data) {
                        if (!data)
                            return;

                        groupSave($item.attr('data-itemid'), {
                            name: data
                        }, SAVE_MODE_CLONE, groupType).then(function () {
                            listUpdate($item.closest('[data-list]'),
                                $item.closest('[data-pane]')
                                    .find('[data-listfilter=' + groupType + ']').val());
                        });
                    }
                },
                labels: {
                    validate: {
                        text: $.t('group.labl_clone')
                    }
                }
            });
        });

        linkRegister('remove-group', function (e) {
            var $item = $(e.target).closest('[data-itemid]');

            overlayCreate('confirm', {
                message: $.t('group.mesg_delete'),
                callbacks: {
                    validate: function () {
                        groupDelete($item.attr('data-itemid'), groupType)
                            .then(function () {
                                listUpdate($item.closest('[data-list]'));
                            })
                            .fail(function () {
                                overlayCreate('alert', {
                                    message: $.t('group.mesg_delete_fail')
                                });
                            });
                    }
                },
                labels: {
                    validate: {
                        text: $.t('group.labl_delete'),
                        style: 'danger'
                    }
                }
            });
        });
    });

    paneRegister('group-edit', function () {
        var groupId = paneMatch('group-edit').opts('pane').id || null,
            groupType = paneMatch('group-edit').opts('pane').section;

        // Register completes and checks
        completionCallbacks = function (input) {
            var $fieldset = input.closest('fieldset');

            if (parseInt($fieldset.find('select[name=type]').val(), 10) !== 0)
                return [];

            return inputGetSources(input, {
                origin: $fieldset.find('input[name=origin]').val()
            });
        };

        if ($('[data-input=source]').length > 0)
            inputRegisterComplete('source', completionCallbacks);

        if ($('[data-input=metric]').length > 0)
            inputRegisterComplete('metric', completionCallbacks);

        if ($('[data-input=group-name]').length > 0) {
            inputRegisterCheck('group-name', function (input) {
                var value = input.find(':input').val();

                if (!value)
                    return;

                groupList({
                    filter: value
                }, groupType).pipe(function (data) {
                    if (data !== null && data[0].id != groupId) {
                        input
                            .attr('title', $.t('group.mesg_exists'))
                            .addClass('error');
                    } else {
                        input
                            .removeAttr('title')
                            .removeClass('error');
                    }
                });
            });
        }

        // Register pane steps
        paneStepRegister('group-edit', 1, function () {
            var $fieldset = $('[data-step=1] fieldset');

            $fieldset.find('button[name=item-update], button[name=item-cancel]').hide();

            if (groupId)
                listSay('step-1-items', null);

            setTimeout(function () { $fieldset.find('input:first').trigger('change').select(); }, 0);
        });

        paneStepRegister('group-edit', 2, function () {
            if (listMatch('step-1-items').find('[data-listitem^=step-1-items-item]').length === 0) {
                overlayCreate('alert', {
                    message: $.t('item.mesg_missing'),
                    callbacks: {
                        validate: function () {
                            setTimeout(function () { $('[data-step=1] fieldset input:first').select(); }, 0);
                        }
                    }
                });
                return false;
            }

            setTimeout(function () { $('[data-step=2] :input:first').select(); });
        });

        // Register links
        linkRegister('remove-item', function (e) {
            var $target = $(e.target),
                $list = $target.closest('[data-list]');

            $target.closest('[data-listitem]').remove();

            listUpdateCount($list);

            if ($list.find('[data-listitem^="' + $list.attr('data-list') + '-item"]').length === 0)
                listSay($list, $.t('item.mesg_none'), 'info');

            PANE_UNLOAD_LOCK = true;
        });

        // Attach events
        $body
            .on('click', 'button', function (e) {
                var $entry,
                    $entryActive,
                    $fieldset,
                    $item,
                    $list,
                    $select,
                    $origin,
                    name,
                    skip = false,
                    type;

                switch (e.target.name) {
                case 'item-add':
                case 'item-update':
                    if (e.target.disabled)
                        return;

                    $fieldset = $(e.target).closest('fieldset');
                    $list     = listMatch('step-1-items');
                    $item     = $fieldset.find('input[name=item]');
                    $origin   = $fieldset.find('select[name=origin]');
                    $select   = $fieldset.find('select[name=type]');

                    if (e.target.name == 'item-update')
                        $entryActive = listMatch('step-1-items').find('[data-listitem^=step-1-items-item].active');

                    type = $select.children('option:selected').text().toLowerCase();

                    $entry = adminGroupCreateItem({
                        pattern: (parseInt($select.val(), 10) !== 0 ? type + ':' : '') + $item.val(),
                        origin: $origin.val()
                    }).find('.type').text(type);

                    if ($entryActive)
                        $entryActive.replaceWith($entry);

                    listSay($list, null);
                    listUpdateCount($list);

                    $item.val('');

                    $item
                        .trigger('change')
                        .focus();

                    PANE_UNLOAD_LOCK = true;

                    break;

                case 'item-cancel':
                    listMatch('step-1-items').find('[data-listitem^=step-1-items-item].active').trigger('click');

                    $(e.target).closest('fieldset').find('input[name=item]')
                        .focus();

                    PANE_UNLOAD_LOCK = true;

                    break;

                case 'step-cancel':
                    window.location = '/admin/' + groupType + '/';
                    break;

                case 'step-save':
                    $(e.target).closest('[data-pane]').find('input[name=group-name]').each(function () {
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

                    groupSave(groupId, adminGroupGetData(), null, groupType)
                        .then(function () {
                            PANE_UNLOAD_LOCK = false;
                            window.location = '/admin/' + groupType + '/';
                        })
                        .fail(function () {
                            overlayCreate('alert', {
                                message: $.t('group.mesg_save_fail')
                            });
                        });

                    break;

                case 'step-ok':
                case 'step-prev':
                case 'step-next':
                    name = $(e.target).closest('[data-pane]').attr('data-pane');

                    if (e.target.name == 'step-ok')
                        paneGoto(name, ADMIN_PANES[name].last);
                    else if (e.target.name == 'step-prev' && ADMIN_PANES[name].active > 1)
                        paneGoto(name, ADMIN_PANES[name].active - 1);
                    else if (e.target.name == 'step-next' && ADMIN_PANES[name].active < ADMIN_PANES[name].count)
                        paneGoto(name, ADMIN_PANES[name].active + 1);

                    break;
                }
            })
            .on('click', '[data-step=1] [data-listitem]', function (e) {
                var $fieldset,
                    $item,
                    $target = $(e.target),
                    active,
                    value;

                if ($target.closest('.actions').length > 0)
                    return;

                $fieldset = $('[data-step=1] fieldset');
                $item     = $target.closest('[data-listitem]');
                value     = $item.data('value');

                $item
                    .toggleClass('active')
                    .siblings().removeClass('active');

                active = $item.hasClass('active');

                if (value.pattern.startsWith('glob:'))
                    value = {origin: value.origin, type: MATCH_TYPE_GLOB, item: value.pattern.substr(5)};
                else if (value.pattern.startsWith('regexp:'))
                    value = {origin: value.origin, type: MATCH_TYPE_REGEXP, item: value.pattern.substr(7)};
                else
                    value = {origin: value.origin, type: MATCH_TYPE_NORMAL, item: value.pattern};

                $fieldset.find('button[name=item-add]').toggle(!active);
                $fieldset.find('button[name=item-update], button[name=item-cancel]').toggle(active);

                $fieldset.find('select[name=origin]')
                    .val(active ? value.origin : '');

                $fieldset.find('input[name=item]')
                    .val(active ? value.item : '');

                $fieldset.find('select[name=type]')
                    .val(active ? value.type : 0)
                    .trigger('change');
            })
            .on('change', '[data-step=1] fieldset input', function (e) {
                var $target = $(e.target),
                    $fieldset = $target.closest('fieldset'),
                    $button = $fieldset.find('button[name=item-add]');

                if (!$fieldset.find('input[name=item]').val())
                    $button.attr('disabled', 'disabled');
                else
                    $button.removeAttr('disabled');

                // Select next item
                if (!e._typing && $target.val())
                    $target.closest('[data-input]').nextAll('button:first').focus();
            })
            .on('change', '[data-step=2] :input', function () {
                PANE_UNLOAD_LOCK = true;
            })
            .on('keyup', '[data-step=1] fieldset input', adminHandleFieldType);

        // Load group data
        if (groupId === null)
            return;

        groupLoad(groupId, groupType).pipe(function (data) {
            var $item,
                $listItems,
                $pane,
                i;

            $listItems = listMatch('step-1-items');

            for (i in data.entries)
                $item = adminGroupCreateItem(data.entries[i]);

            $pane = paneMatch('group-edit');

            $pane.find('input[name=group-name]').val(data.name);
            $pane.find('textarea[name=group-desc]').val(data.description);

            if ($listItems.data('counter') === 0)
                listSay($listItems, $.t('item.mesg_none'));

            listUpdateCount($listItems);
        });
    });
}
