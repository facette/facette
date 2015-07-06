
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

    listGetItems('step-1-items').each(function () {
        data.entries.push($(this).data('value'));
    });

    return data;
}

function adminGroupSetupTerminate() {
    var completionCallbacks;

    // Register admin panes
    paneRegister('group-list', function () {
        adminItemHandlePaneList('group');
    });

    paneRegister('group-edit', function () {
        var groupId = paneMatch('group-edit').opts('pane').id || null,
            groupType = paneMatch('group-edit').opts('pane').section;

        // Register completes and checks
        completionCallbacks = function (input) {
            var $fieldset = input.closest('fieldset');

            if (parseInt($fieldset.find('select[name=type]').val(), 10) != 1)
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

                itemList({
                    filter: value
                }, groupType).pipe(function (data) {
                    if (data.length > 0 && data[0].id != groupId) {
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
            if (listGetCount('step-1-items') === 0) {
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
        linkRegister('move-up move-down', adminItemHandleReorder);

        linkRegister('remove-item', function (e) {
            var $target = $(e.target),
                $list = $target.closest('[data-list]');

            $target.closest('[data-listitem]').remove();

            listUpdateCount($list);

            if (listGetCount($list) === 0)
                listSay($list, $.t('item.mesg_none'), 'info');

            PANE_UNLOAD_LOCK = true;
        });

        linkRegister('test-pattern', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-listitem]');

            $.ajax({
                url: urlPrefix + '/api/v1/catalog/' + (groupType == 'sourcegroups' ? 'sources' : 'metrics') + '/',
                type: 'GET',
                data: {
                    origin: $item.data('value').origin,
                    source: $item.data('value').source,
                    filter: $item.data('value').pattern,
                    limit: PATTERN_TEST_LIMIT
                }
            }).done(function (data, status, xhr) { /*jshint unused: true */
                var $tooltip,
                    records = parseInt(xhr.getResponseHeader('X-Total-Records'), 10);

                $tooltip = tooltipCreate('info', function (state) {
                    $target.toggleClass('active', state);
                    $item.toggleClass('action', state);
                }).appendTo($body)
                    .css({
                        top: $target.offset().top,
                        left: $target.offset().left
                    });

                $tooltip.html('<span class="label">' + $.t('item.labl_matching') + '</span><br>');

                if (data.length === 0) {
                    $tooltip.append($.t('main.mesg_nomatch'));
                } else {
                    $.each(data, function (i, entry) { /*jshint unused: true */
                        $tooltip.append(entry + '<br>');
                    });
                }

                if (records > data.length)
                    $tooltip.append('…<br><span class="label">' + $.t('item.labl_total') + '</span> ' + records);
            });
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
                    type,
                    isPattern;

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
                        $entryActive = listGetItems('step-1-items', '.active');

                    type = $select.children('option:selected').text().toLowerCase();
                    isPattern = parseInt($select.val(), 10) !== MATCH_TYPE_SINGLE;

                    $entry = adminGroupCreateItem({
                        pattern: (isPattern ? type + ':' : '') + $item.val(),
                        origin: $origin.val()
                    });

                    $entry.find('.type').text(type);

                    if (!isPattern)
                        $entry.find('a[href=#test-pattern]').remove();

                    if ($entryActive)
                        $entryActive.replaceWith($entry);

                    listSay($list, null);
                    listUpdateCount($list);

                    $item.val('');

                    $item
                        .trigger('change')
                        .focus();

                    $fieldset.find('button[name=item-add]').show();
                    $fieldset.find('button[name=item-update], button[name=item-cancel]').hide();

                    PANE_UNLOAD_LOCK = true;

                    break;

                case 'item-cancel':
                    listGetItems('step-1-items', '.active').trigger('click');

                    $(e.target).closest('fieldset').find('input[name=item]')
                        .focus();

                    PANE_UNLOAD_LOCK = true;

                    break;

                case 'step-cancel':
                    window.location = urlPrefix + '/admin/' + groupType + '/';
                    break;

                case 'step-save':
                    adminItemHandlePaneSave($(e.target).closest('[data-pane]'), groupId, 'group', adminGroupGetData);
                    break;

                case 'step-ok':
                case 'step-prev':
                case 'step-next':
                    adminHandlePaneStep(e, name);
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
                    value = {origin: value.origin, type: MATCH_TYPE_SINGLE, item: value.pattern};

                $fieldset.find('button[name=item-add]').toggle(!active);
                $fieldset.find('button[name=item-update], button[name=item-cancel]').toggle(active);

                $fieldset.find('select[name=origin]')
                    .val(active ? value.origin : '')
                    .trigger('change');

                $fieldset.find('select[name=type]')
                    .val(active ? value.type : 0)
                    .trigger('change');

                $fieldset.find('input[name=item]')
                    .val(active ? value.item : '');
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
                if (!e._typing && !e._autofill && $target.val())
                    $target.closest('[data-input]').nextAll('button:first').focus();
            })
            .on('change', '[data-step=2] :input', function () {
                PANE_UNLOAD_LOCK = true;
            })
            .on('keyup', '[data-step=1] fieldset input', adminHandleFieldType);

        // Load group data
        if (groupId === null)
            return;

        itemLoad(groupId, groupType).pipe(function (data) {
            var $item,
                $listItems,
                $pane,
                i;

            $listItems = listMatch('step-1-items');

            for (i in data.entries) {
                $item = adminGroupCreateItem(data.entries[i]);

                if (!data.entries[i].pattern.startsWith('glob:') && !data.entries[i].pattern.startsWith('regexp:'))
                    $item.find('a[href=#test-pattern]').remove();
            }

            $pane = paneMatch('group-edit');

            $pane.find('input[name=group-name]').val(data.name);
            $pane.find('textarea[name=group-desc]').val(data.description);

            if ($listItems.data('counter') === 0)
                listSay($listItems, $.t('item.mesg_none'));

            listUpdateCount($listItems);
        });
    });
}
