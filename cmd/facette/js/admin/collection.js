
function adminCollectionCreateGraph(value) {
    var $item = listAppend(listMatch('step-1-graphs'))
        .attr('data-graph', value.id)
        .data('value', value);

    domFillItem($item, value);

    $item.find('.toggle a[href=#hide-options]').hide();
    $item.find('.options').hide();

    return $item;
}

function adminCollectionGetData() {
    var $pane = paneMatch('collection-edit'),
        data = {
            name: $pane.find('input[name=collection-name]').val(),
            description: $pane.find('textarea[name=collection-desc]').val(),
            parent: ($pane.find('input[name=collection-parent]').data('value') || {}).id,
            entries: []
        };

    listGetItems('step-1-graphs').each(function () {
        var $item = $(this),
            $range = $item.find('input[name=graph-range]'),
            $title = $item.find('input[name=graph-title]');

        data.entries.push({
            id: $item.attr('data-graph'),
            options: {
                title: $title.val() || $title.attr('placeholder'),
                range: $range.val() || $range.attr('placeholder'),
                sample: $item.find('input[name=graph-sample]').val(),
                constants: $item.find('input[name=graph-constants]').val(),
                percentiles: $item.find('input[name=graph-percentiles]').val()
            }
        });
    });

    return data;
}

function adminCollectionUpdatePlaceholders(item) {
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
        linkRegister('move-up move-down', function (e) {
            var $target = $(e.target),
                $item = $target.closest('.listitem'),
                $itemNext;

            if (e.target.href.endsWith('#move-up')) {
                $itemNext = $item.prevAll('.listitem:first');

                if ($itemNext.length === 0)
                    return;

                $item.detach().insertBefore($itemNext);
            } else {
                $itemNext = $item.nextAll('.listitem:first');

                if ($itemNext.length === 0)
                    return;

                $item.detach().insertAfter($itemNext);
            }
        });

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

                    adminCollectionUpdatePlaceholders($item);

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
                if (!e._typing && $target.val())
                    $target.closest('[data-input]').nextAll('button:first').focus();
            })
            .on('change', '[data-step=1] .scrollarea :input, [data-step=2] :input', function (e) {
                PANE_UNLOAD_LOCK = true;

                if (e.target.name == 'graph-range')
                    adminCollectionUpdatePlaceholders($(e.target).closest('[data-graph]'));
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
            }

            $pane = paneMatch('collection-edit');

            $pane.find('input[name=collection-name]').val(data.name);
            $pane.find('textarea[name=collection-desc]').val(data.description);

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
                var func = function () {
                        var $item = $(this);

                        domFillItem($item, data[i]);
                        adminCollectionUpdatePlaceholders($item);
                    },
                    i;

                for (i in data)
                    $listGraphs.find('[data-graph=' + data[i].id + ']').each(func);
            });
        });
    });
}
