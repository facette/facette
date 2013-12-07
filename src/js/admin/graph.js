
function adminGraphGetGroup(entry) {
    var $group,
        $listMetrics = listMatch('step-1-metrics'),
        $listOpers = listMatch('step-2-groups'),
        group;

    if (entry.attr('data-group')) {
        $group = $listOpers.find('[data-group=' + entry.attr('data-group') + ']');
        group  = $.extend({series: []}, $group.data('value'));

        $group.find('.groupentry').each(function () {
            group.series.push($.extend({}, $listMetrics.find('[data-serie=' + this.getAttribute('data-serie') +
                ']').data('value')));
        });

        return group;
    } else {
        return {
            name: entry.attr('data-serie'),
            type: OPER_GROUP_TYPE_NONE,
            series: [
                $.extend({}, $listMetrics.find('[data-serie=' + entry.attr('data-serie') + ']').data('value'))
            ]
        };
    }
}

function adminGraphGetStacks() {
    var $listSeries = listMatch('step-stack-series'),
        $listStacks = listMatch('step-stack-groups'),
        groups = [],
        stacks = [];

    // Retrieve defined stacks
    $listStacks.find('[data-listitem^=step-stack-groups-item]').each(function () {
        var $item = $(this);

        groups = [];

        $item.find('.groupentry').each(function () {
            groups.push(adminGraphGetGroup($(this)));
        });

        stacks.push({
            name: $item.attr('data-stack'),
            groups: groups
        });
    });

    // Create new stack with remaining items
    groups = [];

    $listSeries.find('[data-listitem^=step-stack-series-item]:not(.linked)').each(function () {
        groups.push(adminGraphGetGroup($(this)));
    });

    if (groups.length > 0) {
        stacks.push({
            name: listNextName($listStacks, 'data-stack', 'stack'),
            groups: groups
        });
    }

    return stacks;
}

function adminGraphGetData() {
    var $pane = paneMatch('graph-edit'),
        data = {
            name: $pane.find('input[name=graph-name]').val(),
            description: $pane.find('textarea[name=graph-desc]').val(),
            type: parseInt($pane.find('select[name=graph-type]').val(), 10),
            stack_mode: parseInt($pane.find('select[name=stack-mode]').val(), 10),
            stacks: adminGraphGetStacks()
        };

    return data;
}

function adminGraphCreateSerie(value) {
    var $item = listAppend(listMatch('step-1-metrics'))
        .attr('data-serie', value.name)
        .data('proxies', [])
        .data('value', value);

    domFillItem($item, value);

    return $item;
}

function adminGraphCreateGroup(value, list) {
    var $item,
        type;

    list = list || listMatch('step-2-groups');

    if (!value.name)
        value.name = listNextName(list, 'data-group', 'group');

    $item = listAppend(list)
        .attr('data-list', value.name)
        .attr('data-group', value.name)
        .data('proxies', [])
        .data('value', value);

    $item.find('[data-listtmpl]')
        .attr('data-listtmpl', value.name);

    if (value.type == OPER_GROUP_TYPE_AVG)
        type = 'avg';
    else if (value.type == OPER_GROUP_TYPE_SUM)
        type = 'sum';
    else
        type = '';

    $item.find('.name').text(value.name);
    $item.find('.type').text(type);

    return $item;
}

function adminGraphCreateStack(value) {
    var $item;

    if (!value.name)
        value.name = listNextName('step-stack-groups', 'data-stack', 'stack');

    $item = listAppend('step-stack-groups')
        .attr('data-list', value.name)
        .attr('data-stack', value.name)
        .data('value', value);

    $item.find('[data-listtmpl]')
        .attr('data-listtmpl', value.name);

    $item.find('.name').text(value.name);

    return $item;
}

function adminGraphCreateProxy(type, name, list) {
    var $item,
        $itemSrc,
        attr,
        key,
        value;

    if (type !== PROXY_TYPE_SERIE && type !== PROXY_TYPE_GROUP) {
        console.error("Unknown `" + type + "' proxy type");
        return;
    }

    if (type == PROXY_TYPE_SERIE) {
        $itemSrc = listMatch('step-1-metrics').find('[data-serie=' + name + ']')
            .add(listMatch('step-2-series').find('[data-serie=' + name + ']'))
            .first();

        attr = 'data-serie';
    } else {
        $itemSrc = listMatch('step-2-groups').find('[data-group=' + name + ']');

        attr = 'data-group';
    }

    $item = listAppend(list)
        .attr(attr, $itemSrc.attr(attr))
        .data('proxies', []);

    if ($item.attr('data-list') !== undefined)
        $item.attr('data-list', name).find('[data-listtmpl]').attr('data-listtmpl', name);

    $itemSrc.data('proxies').push($item);

    // Copy element values
    value = $itemSrc.data('value');

    for (key in value)
        $item.find('.' + key + ':first').text($itemSrc.find('.' + key + ':first').text());

    return $item;
}

function adminGraphHandleSerieDrag(e) {
    var $group,
        $list,
        $target = $(e.target),
        chunks;

    if (['dragstart', 'dragend'].indexOf(e.type) == -1) {
        $group = $target.closest('.groupitem');

        if ($group.length !== 0) {
            $target = $group;
        } else if (e.target.tagName != 'A' || !$target.attr('href').replace(/\-[^\-]+$/, '').endsWith('#add')) {
            $target = null;
        }
    }

    switch (e.type) {
    case 'dragstart':
        if ($target.hasClass('linked'))
            return;

        $target.addClass('dragged');

        if ($target.attr('data-group'))
            e.dataTransfer.setData('text/plain', 'data-group="' + $target.attr('data-group') + '"');
        else
            e.dataTransfer.setData('text/plain', 'data-serie="' + $target.attr('data-serie') + '"');

        break;

    case 'dragend':
        e.preventDefault();
        $target.removeClass('dragged');
        break;

    case 'dragover':
        e.preventDefault();

        if ($target === null)
            return;

        $target.addClass('dragover');
        e.dataTransfer.dropEffect = 'move';

        break;

    case 'dragleave':
        if ($target === null)
            return;

        $target.removeClass('dragover');

        break;

    case 'drop':
        e.preventDefault();

        if ($target === null)
            return;

        $target.removeClass('dragover');

        // Set item linked
        if (ADMIN_PANES['graph-edit'].active == 'stack')
            $list = listMatch('step-stack-series');
        else
            $list = listMatch('step-2-series');

        $list.find('[' + e.dataTransfer.getData('text/plain') + ']')
            .addClass('linked');

        if ($list.find('[data-listitem^="' + $list.attr('data-list') + '-item"]:not(.linked)').length === 0)
            listSay($list, $.t('graph.mesg_no_serie'));

        // Handle drop'n'create
        if (e.target.tagName == 'A') {
            $target.trigger('click');
            $target = listMatch('step-' + ADMIN_PANES['graph-edit'].active + '-groups').find('.groupitem:last');
        }

        chunks = e.dataTransfer.getData('text/plain').split('=');

        adminGraphCreateProxy(chunks[0] == 'data-group' ? PROXY_TYPE_GROUP : PROXY_TYPE_SERIE, chunks[1],
            $target);

        // Remove item from stack
        if (ADMIN_PANES['graph-edit'].active != 'stack')
            listMatch('step-stack-groups').find('[' + e.dataTransfer.getData('text/plain') + '] a[href=#remove-item]')
                .trigger('click');

        break;
    }
}

function adminGraphSetupTerminate() {
    // Register admin panes
    paneRegister('graph-list', function () {
        // Register links
        linkRegister('edit-graph', function (e) {
            window.location = '/admin/graphs/' + $(e.target).closest('[data-itemid]').attr('data-itemid');
        });

        linkRegister('clone-graph', function (e) {
            var $item = $(e.target).closest('[data-itemid]');

            overlayCreate('prompt', {
                message: $.t('graph.labl_graph_name'),
                value: $item.find('.name').text() + ' (clone)',
                callbacks: {
                    validate: function (data) {
                        if (!data)
                            return;

                        graphSave($item.attr('data-itemid'), {
                            name: data
                        }, SAVE_MODE_CLONE).then(function () {
                            listUpdate($item.closest('[data-list]'),
                                $item.closest('[data-pane]').find('[data-listfilter=graphs]').val());
                        });
                    }
                },
                labels: {
                    validate: {
                        text: $.t('graph.labl_clone')
                    }
                }
            });
        });

        linkRegister('remove-graph', function (e) {
            var $item = $(e.target).closest('[data-itemid]');

            overlayCreate('confirm', {
                message: $.t('graph.mesg_delete'),
                callbacks: {
                    validate: function () {
                        graphDelete($item.attr('data-itemid'))
                            .then(function () {
                                var $list = $item.closest('[data-list]');
                                $item.remove();
                                listUpdateCount($list);
                            })
                            .fail(function () {
                                overlayCreate('alert', {
                                    message: $.t('graph.mesg_delete_fail')
                                });
                            });
                    }
                },
                labels: {
                    validate: {
                        text: $.t('graph.labl_delete'),
                        style: 'danger'
                    }
                }
            });
        });
    });

    paneRegister('graph-edit', function () {
        var graphId = paneMatch('graph-edit').opts('pane').id || null;

        // Register completes and checks
        if ($('[data-input=source]').length > 0) {
            inputRegisterComplete('source', function (input) {
                return inputGetSources(input, {
                    origin: input.closest('fieldset').find('input[name=origin]').val()
                });
            });
        }

        if ($('[data-input=metric]').length > 0) {
            inputRegisterComplete('metric', function (input) {
                var $fieldset = input.closest('fieldset'),
                    source = $fieldset.find('input[name=source]');

                return inputGetSources(input, {
                    origin: $fieldset.find('input[name=origin]').val(),
                    source: (source.data('value').source.endsWith('groups') ? 'group:' : '') + source.val()
                });
            });
        }

        if ($('[data-input=graph-name]').length > 0) {
            inputRegisterCheck('graph-name', function (input) {
                var value = input.find(':input').val();

                if (!value)
                    return;

                graphList({
                    filter: value
                }).pipe(function (data) {
                    if (data !== null && data[0].id != graphId) {
                        input
                            .attr('title', $.t('graph.mesg_exists'))
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
        paneStepRegister('graph-edit', 1, function () {
            var $fieldset = $('[data-step=1] fieldset');

            $fieldset.find('button[name=metric-update], button[name=metric-cancel]').hide();

            if (graphId)
                listSay('step-1-metrics', null);

            setTimeout(function () { $fieldset.find('input:first').trigger('change').select(); }, 0);
        });

        paneStepRegister('graph-edit', 2, function () {
            var $items = listMatch('step-1-metrics').find('[data-listitem^=step-1-metrics-item]'),
                $listOpers,
                $listSeries,
                expand = false,
                query = [];

            if ($items.length === 0) {
                overlayCreate('alert', {
                    message: $.t('metric.mesg_missing'),
                    callbacks: {
                        validate: function () {
                            setTimeout(function () { $('[data-step=1] fieldset input:first').select(); }, 0);
                        }
                    }
                });
                return false;
            }

            // Initialize list
            $listSeries = listMatch('step-2-series');
            $listOpers  = listMatch('step-2-groups');

            listEmpty($listSeries);

            $items.each(function () {
                var $item,
                    $itemSrc = $(this),
                    value;

                value = $itemSrc.data('value');
                query.push([value.origin, value.source, value.metric]);

                if (value.source.startsWith('group:') || value.metric.startsWith('group:'))
                    expand = true;

                $item = adminGraphCreateProxy(PROXY_TYPE_SERIE, value.name, $listSeries);

                if ($listOpers.find('[data-serie=' + value.name + ']').length > 0)
                    $item.addClass('linked');
            });

            if ($listSeries.find('[data-listitem^="step-2-series-item"]:not(.linked)').length > 0)
                listSay($listSeries, null);

            // Retrieve expanding information
            if (expand) {
                $.ajax({
                    url: '/catalog/expand',
                    type: 'POST',
                    contentType: 'application/json',
                    data: JSON.stringify(query)
                }).pipe(function (data) {
                    $listSeries.find('[data-listitem^=step-2-series-item]').each(function (index) {
                        var $item = $(this);

                        if (data[index].length > 1) {
                            $item.data('expand', data[index]);
                            $item.find('.count').text(data[index].length);
                            $item.find('a[href$=#collapse-serie]').remove();
                        } else {
                            $item.find('.count').remove();
                            $item.find('a[href$=#expand-serie], a[href$=#collapse-serie]').remove();
                        }

                        // Restore expanded state
                        if (listMatch('step-1-metrics').find('[data-serie=' + $item.attr('data-serie') +
                                ']').data('expanded'))
                            $item.find('a[href$=#expand-serie]').trigger('click');
                    });
                });
            } else {
                $listSeries.find('.count').remove();
                $listSeries.find('a[href$=#expand-serie], a[href$=#collapse-serie]').remove();
            }
        });

        paneStepRegister('graph-edit', 3, function () {
            var $step = $('[data-step=3]');

            $step.find('select:last').trigger('change');
            setTimeout(function () { $step.find(':input:first').select(); });
        });

        paneStepRegister('graph-edit', 'stack', function () {
            var $item,
                $listSeries,
                $listStacks;

            $listSeries = listMatch('step-stack-series');
            $listStacks = listMatch('step-stack-groups');

            listEmpty($listSeries);

            if (parseInt($('[data-step=3] select[name=stack-mode]').val(), 10) === STACK_MODE_NONE)
                listEmpty($listStacks);

            // Retrieve defined groups
            listMatch('step-2-groups').find('[data-listitem^=step-2-groups-item]').each(function () {
                var $item = adminGraphCreateProxy(PROXY_TYPE_GROUP, this.getAttribute('data-group'), $listSeries);

                if ($listStacks.find('[data-group=' + this.getAttribute('data-group') + ']').length > 0)
                    $item.addClass('linked');

                $(this).find('.groupentry').each(function () {
                    adminGraphCreateProxy(PROXY_TYPE_SERIE, this.getAttribute('data-serie'), $item);
                });
            });

            // Create groups for each remaining items
            listMatch('step-2-series').find('[data-listitem^=step-2-series-item]:not(.linked)').each(function () {
                $item = adminGraphCreateProxy(PROXY_TYPE_SERIE, this.getAttribute('data-serie'), $listSeries);

                if ($listStacks.find('[data-serie=' + this.getAttribute('data-serie') + ']').length > 0)
                    $item.addClass('linked');

                adminGraphCreateProxy(PROXY_TYPE_SERIE, this.getAttribute('data-serie'), $item);
            });

            if ($listSeries.find('[data-listitem^="step-stack-series-item"]:not(.linked)').length > 0)
                listSay($listSeries, null);
        });

        // Register links
        linkRegister('add-avg add-sum', function (e) {
            // Add operation group
            adminGraphCreateGroup({
                type: e.target.href.substr(-3) == 'avg' ? OPER_GROUP_TYPE_AVG : OPER_GROUP_TYPE_SUM
            });

            PANE_UNLOAD_LOCK = true;
        });

        linkRegister('add-stack', function () {
            // Add operation group
            adminGraphCreateStack({});

            PANE_UNLOAD_LOCK = true;
        });

        linkRegister('collapse-serie', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-listitem]'),
                $series,
                collapse,
                name = $item.attr('data-serie').split('-')[0];

            // Unset expansion flag
            listMatch('step-1-metrics').find('[data-serie=' + name + ']').data('expanded', false);

            // Collapse expanded serie
            $series = listMatch('step-2-groups').find('[data-serie^=' + name + '-]');

            collapse = function () {
                $item.siblings('[data-serie=' + name + ']').removeClass('linked');
                $item.siblings('[data-serie^=' + name + '-]').andSelf().remove();

                $series.remove();

                PANE_UNLOAD_LOCK = true;
            };

            if ($series.length > 0) {
                overlayCreate('confirm', {
                    message: $.t('graph.mesg_collapse'),
                    callbacks: {
                        validate: collapse
                    },
                    labels: {
                        validate: {
                            text: $.t('graph.labl_collapse'),
                            style: 'danger'
                        }
                    }
                });
            } else {
                collapse();
            }
        });

        linkRegister('expand-serie', function (e) {
            var $target = $(e.target),
                $item,
                $itemSrc = $target.closest('[data-listitem]'),
                $list = $itemSrc.closest('[data-list]'),
                data = $itemSrc.data('expand'),
                i,
                name = $itemSrc.attr('data-serie'),
                value;

            // Set metric expanded
            listMatch('step-1-metrics').find('[data-serie=' + name + ']').data('expanded', true);

            // Expand serie
            for (i in data) {
                value = {
                    origin: data[i][0],
                    source: data[i][1],
                    metric: data[i][2]
                };

                $item = adminGraphCreateProxy(PROXY_TYPE_SERIE, name, $list)
                    .attr('data-serie', name + '-' + i)
                    .data('value', value);

                domFillItem($item, value);
                $item.find('.name').text($item.attr('data-serie'));

                $item.find('.count').remove();
                $item.find('a[href$=#expand-serie]').remove();

                if (parseInt(i, 10) === 0)
                    $itemSrc.addClass('linked');

                $itemSrc = $item;
            }

            PANE_UNLOAD_LOCK = true;
        });

        linkRegister('remove-group', function (e) {
            var $item,
                $target = $(e.target),
                data,
                i;

            // Remove operation group item
            $item = $target.closest('.groupitem');

            $item.find('.groupentry').each(function () {
                var $entry = $(this),
                    $item;

                if ($entry.attr('data-group'))
                    $item = $target.closest('[data-step]').find('[data-group=' + $entry.attr('data-group') + '].linked')
                        .removeClass('linked');
                else
                    $item = $target.closest('[data-step]').find('[data-serie=' + $entry.attr('data-serie') + '].linked')
                        .removeClass('linked');

                listSay(ADMIN_PANES['graph-edit'].active == 'stack' ? 'step-stack-series' : 'step-2-series', null);
            });

            // Remove proxy items
            data = $item.data('proxies');

            for (i in data)
                data[i].remove();

            $item.remove();

            PANE_UNLOAD_LOCK = true;
        });

        linkRegister('remove-item', function (e) {
            var $target = $(e.target),
                $entry = $target.closest('.groupentry'),
                $item;

            // Remove item from group
            if ($entry.attr('data-group') !== undefined)
                $item = $target.closest('[data-step]').find('[data-group=' + $entry.attr('data-group') + ']')
                    .removeClass('linked');
            else
                $item = $target.closest('[data-step]').find('[data-serie=' + $entry.attr('data-serie') + ']')
                    .removeClass('linked');

            listSay(ADMIN_PANES['graph-edit'].active == 'stack' ? 'step-stack-series' : 'step-2-series', null);

            $target.parent().remove();

            PANE_UNLOAD_LOCK = true;
        });

        linkRegister('remove-metric', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-listitem]'),
                $list = $target.closest('[data-list]'),
                data,
                i;

            // Remove proxy items
            data = $item.data('proxies');

            for (i in data)
                data[i].remove();

            $item.remove();

            listUpdateCount($list);

            if ($list.find('[data-listitem^="' + $list.attr('data-list') + '-item"]').length === 0)
                listSay($list, $.t('metric.mesg_none'), 'info');

            PANE_UNLOAD_LOCK = true;

            $('[data-step=1] fieldset input[name=origin]').focus();
        });

        linkRegister('rename-serie rename-group', function (e) {
            var $target = $(e.target),
                $input,
                $item,
                $overlay,
                attrName,
                value;

            if (e.target.href.endsWith('#rename-group'))
                attrName = 'data-group';
            else
                attrName = 'data-serie';

            $item = $target.closest('[' + attrName + ']');
            value = $item.attr(attrName);

            $overlay = overlayCreate('prompt', {
                message: $.t('graph.labl_serie_name'),
                callbacks: {
                    validate: function (data) {
                        if (!data)
                            return;

                        paneMatch('graph-edit').find('[' + attrName + '="' + value + '"]').each(function () {
                            var $item = $(this),
                                value = $item.data('value');

                            if (value)
                                $item.data('value').name = data;

                            $item
                                .attr(attrName, data)
                                .find('.name').text(data);

                            PANE_UNLOAD_LOCK = true;
                        });
                    }
                }
            });

            $input = $overlay.find('input[type=text]')
                .attr('data-input', 'rename-serie')
                .attr('data-inputopts', 'check: true');

            inputInit($input.get(0));

            $input.val(value);

            inputRegisterCheck('rename-serie', function (input) {
                var valueNew = input.find(':input').val();

                if (valueNew != value && paneMatch('graph-edit').find('[' + attrName + '="' +
                        valueNew + '"]').length > 0) {
                    input
                        .attr('title', $.t('graph.mesg_item_exists'))
                        .addClass('error');

                    $overlay.find('button[name=validate]')
                        .attr('disabled', 'disabled');
                } else {
                    input
                        .removeAttr('title')
                        .removeClass('error');

                    $overlay.find('button[name=validate]')
                        .removeAttr('disabled');
                }
            });
        });

        linkRegister('rename-stack', function (e) {
            var $target = $(e.target),
                $input,
                $item = $target.closest('[data-stack]'),
                $overlay,
                value = $item.attr('data-stack');

            $overlay = overlayCreate('prompt', {
                message: $.t('graph.labl_stack_name'),
                callbacks: {
                    validate: function (value) {
                        if (!value)
                            return;

                        $item
                            .attr('data-stack', value)
                            .find('.name:first').text(value);

                        PANE_UNLOAD_LOCK = true;
                    }
                }
            });

            $input = $overlay.find('input[type=text]')
                .attr('data-input', 'rename-stack')
                .attr('data-inputopts', 'check: true');

            inputInit($input.get(0));

            $input.val(value);

            inputRegisterCheck('rename-stack', function (input) {
                var valueNew = input.find(':input').val();

                if (valueNew != value && paneMatch('graph-edit').find('[data-stack="' + valueNew + '"]').length > 0) {
                    input
                        .attr('title', $.t('graph.mesg_item_exists'))
                        .addClass('error');

                    $overlay.find('button[name=validate]')
                        .attr('disabled', 'disabled');
                } else {
                    input
                        .removeAttr('title')
                        .removeClass('error');

                    $overlay.find('button[name=validate]')
                        .removeAttr('disabled');
                }
            });
        });

        // Attach events
        $body
            .on('click', 'button', function (e) {
                var $fieldset,
                    $input,
                    $item,
                    $itemActive,
                    $list,
                    $metric,
                    $source,
                    $origin,
                    name;

                switch (e.target.name) {
                case 'metric-add':
                case 'metric-update':
                    if (e.target.disabled)
                        return;

                    $list = listMatch('step-1-metrics');

                    $fieldset = $(e.target).closest('fieldset');
                    $metric   = $fieldset.find('input[name=metric]');
                    $source   = $fieldset.find('input[name=source]');
                    $origin   = $fieldset.find('input[name=origin]');

                    if (e.target.name == 'metric-update')
                        $itemActive = listMatch('step-1-metrics').find('[data-listitem^=step-1-metrics-item].active');

                    $item = adminGraphCreateSerie({
                        name: $itemActive && $itemActive.data('value').name ||
                            listNextName('step-1-metrics', 'data-serie', 'serie'),
                        origin: $origin.val(),
                        source: ($source.data('value').source.endsWith('groups') ? 'group:' : '') + $source.val(),
                        metric: ($metric.data('value').source.endsWith('groups') ? 'group:' : '') + $metric.val()
                    });

                    if ($itemActive)
                        $itemActive.replaceWith($item);

                    listSay($list, null);
                    listUpdateCount($list);

                    $origin.val('');
                    $source.val('');
                    $metric.val('');

                    $origin
                        .trigger('change')
                        .focus();

                    PANE_UNLOAD_LOCK = true;

                    break;

                case 'metric-cancel':
                    listMatch('step-1-metrics').find('[data-listitem^=step-1-metrics-item].active').trigger('click');

                    $origin
                        .trigger('change')
                        .focus();

                    PANE_UNLOAD_LOCK = true;

                    break;

                case 'stack-config':
                    paneGoto('graph-edit', 'stack');
                    break;

                case 'step-cancel':
                    window.location = '/admin/graphs/';
                    break;

                case 'step-save':
                    $input = $(e.target).closest('[data-pane]').find('input[name=graph-name]');

                    if (!$input.val()) {
                        $input.closest('[data-input]')
                            .attr('title', $.t('main.mesg_name_missing'))
                            .addClass('error');

                        return;
                    }

                    graphSave(graphId, adminGraphGetData())
                        .then(function () {
                            PANE_UNLOAD_LOCK = false;
                            window.location = '/admin/graphs/';
                        })
                        .fail(function () {
                            overlayCreate('alert', {
                                message: $.t('graph.mesg_save_fail')
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

                $fieldset.find('button[name=metric-add]').toggle(!active);
                $fieldset.find('button[name=metric-update], button[name=metric-cancel]').toggle(active);

                $fieldset.find('input[name=origin]').val(active ? value.origin : '');
                $fieldset.find('input[name=source]').val(active ? value.source : '');
                $fieldset.find('input[name=metric]').val(active ? value.metric : '').trigger('change');
            })
            .on('change', '[data-step=1] fieldset input', function (e) {
                var $buttons,
                    $target = $(e.target),
                    $fieldset = $target.closest('fieldset'),
                    $next;

                if (!$fieldset.find('input[name=origin]').val())
                    $fieldset.find('input[name=source]')
                        .val('')
                        .attr('disabled', 'disabled');
                else
                    $fieldset.find('input[name=source]')
                        .removeAttr('disabled');

                if (!$fieldset.find('input[name=source]').val())
                    $fieldset.find('input[name=metric]')
                        .val('')
                        .attr('disabled', 'disabled');
                else
                    $fieldset.find('input[name=metric]')
                        .removeAttr('disabled');

                $buttons = $fieldset.find('button[name=metric-add], button[name=metric-update]');

                if (!$fieldset.find('input[name=origin]').val() || !$fieldset.find('input[name=source]').val() ||
                        !$fieldset.find('input[name=metric]').val()) {
                    $buttons.attr('disabled', 'disabled');
                } else {
                    $buttons.removeAttr('disabled');
                }

                // Select next item
                if ($target.val()) {
                    $next = $target.closest('[data-input]').nextAll('[data-input], button:visible').first();

                    if ($next.attr('data-input') !== undefined)
                        $next = $next.children('input');

                    $next.focus();
                }
            })
            .on('change', '[data-step=3] select', function (e) {
                var $target = $(e.target);

                if (!e._select)
                    return;

                if (e.target.name == 'stack-mode') {
                    $target.closest('[data-step]').find('button[name=stack-config]')
                        .toggle(parseInt(e.target.value, 10) !== STACK_MODE_NONE);

                    paneGoto('graph-edit', 'stack', true);
                }

                graphSave(null, adminGraphGetData(), SAVE_MODE_VOLATILE).pipe(function (data, status, xhr) {
                    /*jshint unused: true */
                    var location = xhr.getResponseHeader('Location');

                    graphDraw($target.closest('[data-step]').find('[data-graph]').attr('data-graph',
                        location.substr(location.lastIndexOf('/') + 1)));
                });
            })
            .on('change', '[data-step=3] :input', function () {
                PANE_UNLOAD_LOCK = true;
            })
            .on('keydown', '[data-step=1] fieldset input', function (e) {
                $(e.target).nextAll('input')
                    .attr('disabled', 'disabled')
                    .val('');
            })
            .on('dragstart dragend dragover dragleave drop', '.dragarea', adminGraphHandleSerieDrag);

        // Load graph data
        if (graphId === null)
            return;

        graphLoad(graphId).pipe(function (data) {
            var $itemOper,
                $itemSerie,
                $itemStack,
                $listMetrics,
                $listOpers,
                $pane,
                i,
                j,
                k;

            $listMetrics = listMatch('step-1-metrics');
            $listOpers   = listMatch('step-2-groups');

            for (i in data.stacks) {
                $itemStack = data.stacks[i].mode !== STACK_MODE_NONE ? adminGraphCreateStack({
                   name: data.stacks[i].name
                }) : null;

                for (j in data.stacks[i].groups) {
                    $itemOper = data.stacks[i].groups[j].type !== OPER_GROUP_TYPE_NONE ? adminGraphCreateGroup({
                     name: data.stacks[i].groups[j].name,
                     type: data.stacks[i].groups[j].type
                    }) : null;

                    for (k in data.stacks[i].groups[j].series) {
                        $itemSerie = adminGraphCreateSerie(data.stacks[i].groups[j].series[k]);

                        if ($itemOper)
                            adminGraphCreateProxy(PROXY_TYPE_SERIE, data.stacks[i].groups[j].series[k].name,
                                $itemOper);
                        else if ($itemStack)
                            adminGraphCreateProxy(PROXY_TYPE_SERIE, data.stacks[i].groups[j].series[k].name,
                                $itemStack);
                    }

                    if ($itemOper && $itemStack)
                        adminGraphCreateProxy(PROXY_TYPE_GROUP, data.stacks[i].groups[j].name, $itemStack);
                }
            }

            $pane = paneMatch('graph-edit');

            $pane.find('input[name=graph-name]').val(data.name);
            $pane.find('textarea[name=graph-desc]').val(data.description);
            $pane.find('select[name=graph-type]').val(data.type);
            $pane.find('select[name=stack-mode]').val(data.stack_mode);

            if ($listMetrics.data('counter') === 0)
                listSay($listMetrics, $.t('metric.mesg_none'));

            listUpdateCount($listMetrics);
        });
    });
}
