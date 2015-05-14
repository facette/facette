
function adminCollectionCreateGraph(value) {
    var $list = listMatch('step-1-graphs'),
        $item,
        id;

    $item = listAppend($list)
        .attr('data-graph', value.id)
        .data('value', value);

    domFillItem($item, value);

    $item.find('.toggle a[href=#hide-options]').hide();
    $item.find('.options').hide();

    // Make checkbox id unique
    id = 'graph-shown-item' + (listGetCount($list) - 1);

    $item.find('input[name=graph-shown]')
        .attr('id', id);
    $item.find('label[for=graph-shown]')
        .attr('for', id);

    return $item;
}

function adminCollectionGetData() {
    var $pane = paneMatch('collection-edit'),
        data = {
            name: $pane.find('input[name=collection-name]').val(),
            description: $pane.find('textarea[name=collection-desc]').val(),
            parent: ($pane.find('input[name=collection-parent]').data('value') || {id: null}).id,
            entries: []
        },
        refresh_interval = $pane.find('input[name=collection-refresh-interval]').val();

    if (refresh_interval)
        data.options = {refresh_interval: parseInt(refresh_interval, 10)};

    listGetItems('step-1-graphs').each(function () {
        var $item = $(this),
            $range = $item.find('input[name=graph-range]'),
            $title = $item.find('input[name=graph-title]'),
            options,
            value;

        options = {
            title: $title.val() || null,
            range: $range.val() || null,
            constants: $item.find('input[name=graph-constants]').val(),
            percentiles: $item.find('input[name=graph-percentiles]').val(),
            enabled: $item.find('input[name=graph-shown]').is(':checked')
        };

        options.constants = parseFloatList(options.constants);
        options.percentiles = parseFloatList(options.percentiles);

        value = $item.find('input[name=graph-sample]').val();
        if (value)
            options.sample = parseInt(value, 10);

        value = $item.find('input[name=graph-refresh-interval]').val();
        if (value)
            options.refresh_interval = parseInt(value, 10);

        data.entries.push({
            id: $item.attr('data-graph'),
            options: options
        });
    });

    return data;
}

function adminCollectionUpdatePlaceholder(item) {
    item.find('input[name=graph-title]').attr('placeholder', item.find('.name').text());
}

function adminCollectionSetupTerminate() {
    // Register admin panes
    paneRegister('collection-list', function () {
        adminItemHandlePaneList('collection');
    });

    paneRegister('collection-edit', function () {
        var collectionId = paneMatch('collection-edit').opts('pane').id || null;

        // Register completes and checks
        if ($('[data-input=graph]').length > 0) {
            inputRegisterComplete('graph', function (input) {
                return inputGetSources(input, {});
            });
        }

        if ($('[data-input=collection]').length > 0) {
            inputRegisterComplete('collection', function (input) {
                return inputGetSources(input, {
                    exclude: $(input).opts('input').exclude
                });
            });
        }

        if ($('[data-input=collection-name]').length > 0) {
            inputRegisterCheck('collection-name', function (input) {
                var value = input.find(':input').val();

                if (!value)
                    return;

                itemList({
                    filter: value
                }, 'collections').pipe(function (data) {
                    if (data.length > 0 && data[0].id != collectionId) {
                        input
                            .attr('title', $.t('collection.mesg_exists'))
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
        paneStepRegister('collection-edit', 1, function () {
            if (collectionId)
                listSay('step-1-graphs', null);

            setTimeout(function () { $('[data-step=1] fieldset input:first').trigger('change').select(); }, 0);
        });

        paneStepRegister('collection-edit', 2, function () {
            setTimeout(function () { $('[data-step=2] :input:first').select(); });
        });

        // Register links
        linkRegister('edit-graph', function (e) {
            var $item = $(e.target).closest('[data-graph]'),
                location;

            location = urlPrefix + '/admin/graphs/' + $item.attr('data-graph');
            if ($item.data('params'))
                location += '?' + $item.data('params');

            window.location = location;
        });

        linkRegister('move-up move-down', adminItemHandleReorder);

        linkRegister('remove-graph', function (e) {
            var $target = $(e.target),
                $list = $target.closest('[data-list]');

            $target.closest('[data-listitem]').remove();

            listUpdateCount($list);

            if (listGetCount($list) === 0)
                listSay($list, $.t('graph.mesg_none'), 'info');

            PANE_UNLOAD_LOCK = true;
        });

        linkRegister('show-options hide-options', function (e) {
            var $item = $(e.target).closest('[data-listitem]');

            $item.find('.toggle a').toggle();
            $item.find('.options').toggle();
        });

        // Attach events
        $body
            .on('click', 'button', function (e) {
                var $fieldset,
                    $graph,
                    $item,
                    $list,
                    $target = $(e.target),
                    name;

                switch (e.target.name) {
                case 'graph-add':
                    if (e.target.disabled)
                        return;

                    $fieldset = $target.closest('fieldset');
                    $list     = listMatch('step-1-graphs');
                    $graph    = $fieldset.find('input[name=graph]');

                    if (!$graph.data('value')) {
                        if (!e._retry) {
                            $.ajax({
                                url: urlPrefix + '/api/v1/library/graphs/',
                                type: 'GET',
                                data: {
                                    filter: $graph.val()
                                },
                                dataType: 'json'
                            }).pipe(function (data) {
                                if (data)
                                    $graph.data('value', data[0]);

                                $target.trigger({
                                    type: 'click',
                                    _retry: true
                                });
                            });

                            return;
                        }

                        overlayCreate('alert', {
                            message: $.t('graph.mesg_unknown'),
                            callbacks: {
                                validate: function () {
                                    setTimeout(function () { $graph.select(); }, 0);
                                }
                            }
                        });

                        return;
                    }

                    $item = adminCollectionCreateGraph({
                        id: $graph.data('value').id,
                        name: $graph.val()
                    });

                    if ($graph.data('value').link)
                        $item.data('params', 'linked=1');

                    $item.find('input[name=graph-shown]').attr('checked', 'checked');

                    adminCollectionUpdatePlaceholder($item);

                    listSay($list, null);
                    listUpdateCount($list);

                    $graph.val('');

                    $graph
                        .trigger('change')
                        .focus();

                    PANE_UNLOAD_LOCK = true;

                    break;

                case 'step-cancel':
                    window.location = urlPrefix + '/admin/collections/';
                    break;

                case 'step-save':
                    adminItemHandlePaneSave($target.closest('[data-pane]'), collectionId, 'collection',
                        adminCollectionGetData);
                    break;

                case 'step-ok':
                case 'step-prev':
                case 'step-next':
                    adminHandlePaneStep(e, name);
                    break;
                }
            })
            .on('change', '[data-step=1] fieldset input', function (e) {
                var $target = $(e.target),
                    $fieldset = $target.closest('fieldset'),
                    $button = $fieldset.find('button[name=graph-add]');

                if (!$fieldset.find('input[name=graph]').val())
                    $button.attr('disabled', 'disabled');
                else
                    $button.removeAttr('disabled');

                // Select next item
                if (!e._typing && !e._autofill && $target.val())
                    $target.closest('[data-input]').nextAll('button:first').focus();
            })
            .on('change', '[data-step=1] .scrollarea input[type=checkbox]', function (e) {
                var $target = $(e.target);
                $target.closest('[data-listitem]').toggleClass('hidden', !$target.is(':checked'));
            })
            .on('change', '[data-step=1] .scrollarea :input, [data-step=2] :input', function (e) {
                PANE_UNLOAD_LOCK = true;

                if (e.target.name == 'graph-range')
                    adminCollectionUpdatePlaceholder($(e.target).closest('[data-graph]'));
            })
            .on('keyup', '[data-step=1] fieldset input', adminHandleFieldType);

        // Load collection data
        if (collectionId === null)
            return;

        itemLoad(collectionId, 'collections').pipe(function (data) {
            var $item,
                $listGraphs,
                $pane,
                i,
                query = {};

            $listGraphs = listMatch('step-1-graphs');

            for (i in data.entries) {
                $item = adminCollectionCreateGraph(data.entries[i]);
                $item.find('input[name=graph-title]').val(data.entries[i].options.title || '');
                $item.find('input[name=graph-range]').val(data.entries[i].options.range || '');
                $item.find('input[name=graph-sample]').val(data.entries[i].options.sample || '');
                $item.find('input[name=graph-constants]').val(data.entries[i].options.constants || '');
                $item.find('input[name=graph-percentiles]').val(data.entries[i].options.percentiles || '');

                if (data.entries[i].options.enabled)
                    $item.find('input[name=graph-shown]').attr('checked', 'checked');
                else
                    $item.addClass('hidden');

                if (data.entries[i].options.refresh_interval)
                    $item.find('input[name=graph-refresh-interval]').val(data.entries[i].options.refresh_interval);
            }

            $pane = paneMatch('collection-edit');

            $pane.find('input[name=collection-name]').val(data.name);
            $pane.find('textarea[name=collection-desc]').val(data.description);

            if (data.options && data.options.refresh_interval)
                $pane.find('input[name=collection-refresh-interval]').val(data.options.refresh_interval);

            if (data.parent) {
                itemLoad(data.parent, 'collections').pipe(function (data) {
                    $pane.find('input[name=collection-parent]')
                        .data('value', data)
                        .val(data.name);
                });
            }

            if ($listGraphs.data('counter') === 0)
                listSay($listGraphs, $.t('graph.mesg_none'));

            listUpdateCount($listGraphs);

            // Load missing graph data
            if (collectionId)
                query.collection = collectionId;

            itemList(query, 'graphs').pipe(function (data) {
                var info = {},
                    i;

                for (i in data)
                    info[data[i].id] = data[i];

                $listGraphs.find('[data-graph]').each(function () {
                    var $item = $(this),
                        id = $item.attr('data-graph');

                    if (!info[id]) {
                        $item.addClass('unknown');
                        info[id] = {name: $.t('graph.mesg_unknown')};
                    }

                    if (info[id].link)
                        $item.data('params', 'linked=1');

                    domFillItem($item, info[id]);
                    adminCollectionUpdatePlaceholder($item);
                });
            });
        });
    });
}
