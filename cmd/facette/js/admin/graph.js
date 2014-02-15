
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

function adminGraphGetGroup(entry) {
    var group;

    if (entry.attr('data-group')) {
        group = $.extend({
            series: [],
            options: {}
        }, adminGraphGetValue(entry));

        listMatch('step-2-groups').find('[data-group="' + entry.attr('data-group') + '"] .groupentry')
            .each(function () {
                group.series.push(adminGraphGetValue($(this)));
            });
    } else {
        group = {
            name: entry.attr('data-serie'),
            type: OPER_GROUP_TYPE_NONE,
            series: [],
            options: {}
        };

        group.series.push(adminGraphGetValue(entry));

        if (group.series[0])
            group.options = group.series[0].options || {};
    }

    return group;
}

function adminGraphGetStacks() {
    var $listSeries = listMatch('step-stack-series'),
        $listStacks = listMatch('step-stack-groups'),
        groups = [],
        stacks = [];

    // Retrieve defined stacks
    listGetItems($listStacks).each(function () {
        var $item = $(this);

        groups = [];

        $item.find('.groupentry').each(function () {
            groups.push(adminGraphGetGroup($(this)));
        });

        if (groups.length === 0)
            return;

        stacks.push({
            name: $item.attr('data-stack'),
            groups: groups
        });
    });

    // Create new stack with remaining items
    groups = [];

    listGetItems($listSeries, ':not(.linked)').each(function () {
        groups.push(adminGraphGetGroup($(this)));
    });

    if (groups.length > 0) {
        stacks.push({
            name: listNextName($listStacks, 'data-stack'),
            groups: groups
        });
    }

    return stacks;
}

function adminGraphGetValue(item) {
    if (item.data('source')) {
        if (item.hasClass('expand'))
            return item.data('source').data('expands')[item.attr('data-serie')];
        else
            return item.data('source').data('value');
    } else {
        return item.data('value');
    }
}

function adminGraphCreateSerie(name, value) {
    var $item,
        $list = listMatch('step-1-metrics');

    // Set defaults
    if (!name)
        name = listNextName($list, 'data-serie');

    if (!value.name)
        value.name = name;

    // Create new serie
    $item = listAppend($list)
        .attr('data-serie', name)
        .data({
            expands: {},
            proxies: [],
            value: value
        });

    domFillItem($item, value);

    return $item;
}

function adminGraphCreateGroup(name, value) {
    var $item,
        $list = listMatch('step-2-groups'),
        type;

    // Set defaults
    if (!name)
        name = listNextName($list, 'data-group');

    if (!value.name)
        value.name = name;

    // Create new group
    $item = listAppend($list)
        .attr({
            'data-group': name,
            'data-list': name
        })
        .data({
            proxies: [],
            value: value
        });

    $item.find('[data-listtmpl]')
        .attr('data-listtmpl', name);

    if (value.type == OPER_GROUP_TYPE_AVG)
        type = 'avg';
    else if (value.type == OPER_GROUP_TYPE_SUM)
        type = 'sum';
    else
        type = '';

    // Update group
    domFillItem($item, {
        name: value.name,
        type: type
    });

    if (value.options && value.options.color)
        $item.find('.color')
            .removeClass('auto')
            .css('color', value.options.color);

    if (value.scale)
        $item.find('a[href=#set-scale]').text(value.scale.toPrecision(3));

    return $item;
}

function adminGraphCreateStack(value) {
    var $item,
        $list = listMatch('step-stack-groups'),
        name;

    // Set defaults
    name = listNextName($list, 'data-stack');

    if (!value.name)
        value.name = name;

    $item = listAppend($list)
        .attr({
            'data-stack': value.name,
            'data-list': value.name
        })
        .data('value', value);

    $item.find('[data-listtmpl]')
        .attr('data-listtmpl', name);

    domFillItem($item, {
        name: value.name,
    });

    return $item;
}

function adminGraphCreateProxy(type, item, list) {
    var $item,
        attr,
        name,
        value;

    switch (type) {
    case PROXY_TYPE_SERIE:
        attr = 'data-serie';
        break;

    case PROXY_TYPE_GROUP:
        attr = 'data-group';
        break;

    default:
        console.error("Unknown `" + type + "' proxy type");
        return;
    }

    name = item.attr(attr);

    $item = listAppend(list)
        .attr(attr, name)
        .data('source', item);

    if ($item.attr('data-list') !== undefined)
        $item.attr('data-list', name).find('[data-listtmpl]').attr('data-listtmpl', name);

    item.data('proxies').push($item);

    // Update proxy
    value = $.extend({}, item.data('value'));

    if (type == PROXY_TYPE_GROUP) {
        if (value.type == OPER_GROUP_TYPE_AVG)
            value.type = 'avg';
        else if (value.type == OPER_GROUP_TYPE_SUM)
            value.type = 'sum';
        else
            delete value.type;
    }

    domFillItem($item, value);

    if (value.options && value.options.color)
        $item.find('.color')
            .removeClass('auto')
            .css('color', value.options.color);

    if (value.scale)
        $item.find('a[href=#set-scale]').text(value.scale.toPrecision(3));

    return $item;
}

function adminGraphHandleSerieDrag(e) {
    var $group,
        $item,
        $itemSrc,
        $listSeries,
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
        if ($target.hasClass('linked') || !$target.attr('data-serie') && !$target.attr('data-group'))
            return;

        $target.addClass('dragged');

        if ($target.attr('data-group'))
            e.dataTransfer.setData('text/plain', 'data-group=' + $target.attr('data-group'));
        else
            e.dataTransfer.setData('text/plain', 'data-serie=' + $target.attr('data-serie'));

        break;

    case 'dragend':
        e.preventDefault();
        $target.removeClass('dragged');
        break;

    case 'dragover':
        e.preventDefault();

        if ($target === null || !e.dataTransfer.getData('text/plain').startsWith('data-'))
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
            $listSeries = listMatch('step-stack-series');
        else
            $listSeries = listMatch('step-2-series');

        $listSeries.find('[' + e.dataTransfer.getData('text/plain') + ']')
            .addClass('linked');

        if (listGetCount($listSeries, ':not(.linked)') === 0)
            listSay($listSeries, $.t('graph.mesg_no_serie'));

        // Handle drop'n'create
        if (e.target.tagName == 'A') {
            $target.trigger('click');
            $target = listMatch('step-' + ADMIN_PANES['graph-edit'].active + '-groups').find('.groupitem:last');
        }

        chunks = e.dataTransfer.getData('text/plain').split('=');

        if (chunks[0] == 'data-serie') {
            $itemSrc = listMatch('step-2-series').find('[' + e.dataTransfer.getData('text/plain') + ']');

            $item = adminGraphCreateProxy(PROXY_TYPE_SERIE, $itemSrc.data('source'), $target);

            if ($itemSrc.hasClass('expand')) {
                $item.attr('data-serie', chunks[1]);
                domFillItem($item, $itemSrc.data('source').data('expands')[chunks[1]]);
            }
        } else {
            $itemSrc = listMatch('step-stack-series').find('[' + e.dataTransfer.getData('text/plain') + ']');

            adminGraphCreateProxy(PROXY_TYPE_GROUP, $itemSrc.data('source'), $target);
        }

        // Remove item from stack
        if (ADMIN_PANES['graph-edit'].active != 'stack')
            listMatch('step-stack-groups').find('[' + e.dataTransfer.getData('text/plain') + '] a[href=#remove-item]')
                .trigger('click');

        break;
    }
}

function adminGraphRestorePosition() {
    var $parent = null,
        items = [];

    listGetItems('step-2-series').each(function () {
        var $item = $(this);

        if (!$parent)
            $parent = $item.parent();

        items.push([$item.detach(), adminGraphGetValue($item).position]);
    });

    items.sort(function (x, y) {
        return x[1] - y[1];
    });

    $.each(items, function (i, item) { /*jshint unused: true */
        item[0].appendTo($parent);
    });
}

function adminGraphSetupTerminate() {
    // Register admin panes
    paneRegister('graph-list', function () {
        adminItemHandlePaneList('graph');
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
                    source: (source.data('value') && source.data('value').source.endsWith('groups') ? 'group:' : '') +
                        source.val()
                });
            });
        }

        if ($('[data-input=graph-name]').length > 0) {
            inputRegisterCheck('graph-name', function (input) {
                var value = input.find(':input').val();

                if (!value)
                    return;

                itemList({
                    filter: value
                }, 'graphs').pipe(function (data) {
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
            var $items = listGetItems('step-1-metrics'),
                $listOpers,
                $listSeries,
                expand = false,
                expandQuery = [];

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
                    value = $itemSrc.data('value');

                if (value.source.startsWith('group:') || value.metric.startsWith('group:')) {
                    expandQuery.push([value.origin, value.source, value.metric]);
                    expand = true;
                }

                $item = adminGraphCreateProxy(PROXY_TYPE_SERIE, $itemSrc, $listSeries);

                if ($listOpers.find('[data-serie="' + $itemSrc.attr('data-serie') + '"]').length > 0)
                    $item.addClass('linked');
            });

            adminGraphRestorePosition();

            if (listGetCount($listSeries, ':not(.linked)') > 0)
                listSay($listSeries, null);

            // Retrieve expanding information
            if (expand) {
                $.ajax({
                    url: urlPrefix + '/library/expand',
                    type: 'POST',
                    contentType: 'application/json',
                    data: JSON.stringify(expandQuery)
                }).pipe(function (data) {
                    listGetItems($listSeries).each(function (index) {
                        var $item = $(this);

                        if (data[index]) {
                            $item.find('.count').text(data[index].length);

                            if (data[index].length > 1) {
                                $item.data('expand', data[index]);
                                $item.find('a[href$=#collapse-serie]').remove();
                            }
                        } else {
                            $item.find('.count').remove();
                            $item.find('a[href$=#expand-serie], a[href$=#collapse-serie]').remove();
                        }

                        // Restore expanded state
                        if (!$.isEmptyObject(listMatch('step-1-metrics').find('[data-serie=' +
                                $item.attr('data-serie') + ']').data('expands')))
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
            var $listSeries = listMatch('step-stack-series'),
                $listStacks = listMatch('step-stack-groups');

            listEmpty($listSeries);

            if (parseInt($('[data-step=3] select[name=stack-mode]').val(), 10) === STACK_MODE_NONE)
                listEmpty($listStacks);

            // Retrieve defined groups
            listGetItems('step-2-groups').each(function () {
                var $item,
                    $itemSrc = $(this);

                $item = adminGraphCreateProxy(PROXY_TYPE_GROUP, $itemSrc, $listSeries);

                if ($listStacks.find('[data-group="' + $itemSrc.attr('data-group') + '"]').length > 0)
                    $item.addClass('linked');

                $itemSrc.find('.groupentry').each(function () {
                    adminGraphCreateProxy(PROXY_TYPE_SERIE, $(this).data('source'), $item);
                });
            });

            // Create groups for each remaining items
            listGetItems('step-2-series').each(function () {
                var $item,
                    $itemMain,
                    $itemSrc = $(this),
                    value;

                if ($itemSrc.hasClass('linked')) {
                    $listStacks.find('[data-serie="' + $itemSrc.attr('data-serie') + '"]').remove();
                    return;
                }

                $itemMain = adminGraphCreateProxy(PROXY_TYPE_SERIE, $itemSrc.data('source'), $listSeries)
                    .attr('data-serie', $itemSrc.attr('data-serie'));

                if ($listStacks.find('[data-serie="' + $itemSrc.attr('data-serie') + '"]').length > 0)
                    $itemMain.addClass('linked');

                $item = adminGraphCreateProxy(PROXY_TYPE_SERIE, $itemSrc.data('source'), $itemMain);

                if ($itemSrc.hasClass('expand')) {
                    value = $itemSrc.data('source').data('expands')[$itemSrc.attr('data-serie')];
                    $itemMain.find('.name:first').text(value.name);
                } else {
                    value = $itemSrc.data('source').data('value');
                }

                domFillItem($item, value);

                if ($itemSrc.hasClass('expand'))
                    $itemMain.addClass('expand');
            });

            if (listGetCount($listSeries, ':not(.linked)') > 0)
                listSay($listSeries, null);
        });

        // Register links
        linkRegister('add-avg add-sum add-stack', function (e) {
            if (e.target.href.substr(-5) == 'stack') {
                // Add stack group
                adminGraphCreateStack({});
            } else {
                // Add operation group
                adminGraphCreateGroup(null, {
                    type: e.target.href.substr(-3) == 'avg' ? OPER_GROUP_TYPE_AVG : OPER_GROUP_TYPE_SUM
                });
            }

            PANE_UNLOAD_LOCK = true;
        });

        linkRegister('collapse-serie', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-serie]'),
                $series,
                collapse,
                name = $item.attr('data-serie').split('-')[0];

            // Collapse expanded serie
            $series = listMatch('step-2-groups').find('[data-serie^="' + name + '-"]');

            collapse = function () {
                $item.data('source').data('expands', {});

                $item.siblings('[data-serie="' + name + '"]').removeClass('linked');
                $item.siblings('[data-serie^="' + name + '-"]').andSelf().remove();

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
                $itemSrc = $target.closest('[data-serie]'),
                $itemRef = $itemSrc,
                $listSeries = $itemSrc.closest('[data-list]'),
                data = $itemSrc.data('expand'),
                expands = $itemSrc.data('source').data('expands'),
                i,
                name = $itemSrc.attr('data-serie'),
                serieName;

            // Expand serie
            for (i in data) {
                serieName = name + '-' + i;

                $item = adminGraphCreateProxy(PROXY_TYPE_SERIE, $itemSrc.data('source'), $listSeries)
                    .attr('data-serie', serieName)
                    .addClass('expand');

                $item.detach().insertAfter($itemRef);

                if (!expands[serieName]) {
                    expands[serieName] = {
                        name: serieName,
                        origin: data[i][0],
                        source: data[i][1],
                        metric: data[i][2]
                    };
                }

                domFillItem($item, expands[serieName]);

                if (expands[serieName].options && expands[serieName].options.color)
                    $item.find('.color')
                        .removeClass('auto')
                        .css('color', expands[serieName].options.color);

                if (expands[serieName].scale !== 0)
                    $item.find('a[href=#set-scale]').text(expands[serieName].scale);

                $item.find('.count').remove();
                $item.find('a[href$=#expand-serie]').remove();

                if (parseInt(i, 10) === 0)
                    $itemSrc.addClass('linked');

                $itemRef = $item;
            }

            adminGraphRestorePosition();

            PANE_UNLOAD_LOCK = true;
        });

        linkRegister('move-up move-down', function (e) {
            var $target = $(e.target),
                $item = $target.closest('.listitem, .groupitem, .groupentry'),
                $itemNext;

            if (e.target.href.endsWith('#move-up')) {
                $itemNext = $item.prevAll('.listitem, .groupitem, .groupentry').filter(':not(.linked):last');

                if ($itemNext.length === 0)
                    return;

                $item.detach().insertBefore($itemNext);
            } else {
                $itemNext = $item.nextAll('.listitem, .groupitem, .groupentry').filter(':not(.linked):first');

                if ($itemNext.length === 0)
                    return;

                $item.detach().insertAfter($itemNext);
            }

            // Save positions
            if (ADMIN_PANES['graph-edit'].active == 'stack' || !$item.hasClass('listitem'))
                return;

            listGetItems('step-2-series').each(function () {
                var $item = $(this);
                adminGraphGetValue($item).position = $item.index();
            });
        });

        linkRegister('remove-group', function (e) {
            var $item,
                $target = $(e.target),
                data,
                i;

            // Remove operation group item
            $item = $target.closest('.groupitem');
            $item.find('.groupentry a[href=#remove-item]').trigger('click');

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
            if ($entry.attr('data-group'))
                $item = $target.closest('[data-step]').find('[data-group="' + $entry.attr('data-group') + '"]')
                    .removeClass('linked');
            else
                $item = $target.closest('[data-step]').find('[data-serie="' + $entry.attr('data-serie') + '"]')
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

            if (listGetCount($list) === 0)
                listSay($list, $.t('metric.mesg_none'), 'info');

            PANE_UNLOAD_LOCK = true;

            $('[data-step=1] fieldset input[name=origin]').focus();
        });

        linkRegister('rename-serie rename-group rename-stack', function (e) {
            var $target = $(e.target),
                $input,
                $item,
                $overlay,
                attrName,
                serieName,
                value;

            if (e.target.href.endsWith('#rename-stack'))
                attrName = 'data-stack';
            else if (e.target.href.endsWith('#rename-group'))
                attrName = 'data-group';
            else
                attrName = 'data-serie';

            $item     = $target.closest('[' + attrName + ']');
            serieName = $item.attr(attrName);

            value = adminGraphGetValue($item).name;

            $overlay = overlayCreate('prompt', {
                message: $.t(e.target.href.endsWith('#rename-stack') ? 'graph.labl_stack_name' :
                    'graph.labl_serie_name'),
                callbacks: {
                    validate: function (data) {
                        if (!data)
                            return;

                        adminGraphGetValue($item).name = data;

                        paneMatch('graph-edit').find('[' + attrName + '="' + serieName + '"]').each(function () {
                            $(this).find('.name:first').text(data);
                        });

                        PANE_UNLOAD_LOCK = true;
                    }
                }
            });

            $input = $overlay.find('input[type=text]')
                .attr({
                    'data-input': 'rename-item',
                    'data-inputopts': 'check: true'
                });

            inputInit($input.get(0));

            $input.val(value);

            inputRegisterCheck('rename-item', function (input) {
                var valueNew = input.find(':input').val(),
                    values = [];

                listGetItems('step-1-metrics').add(listGetItems('step-2-groups')).each(function () {
                    var $item = $(this),
                        serieName;

                    values.push($(this).data('value').name);

                    if (!$item.data('expands'))
                        return;

                    for (serieName in $item.data('expands'))
                        values.push($item.data('expands')[serieName].name);
                });

                if (values.indexOf(valueNew) != -1) {
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

        linkRegister('set-color', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-serie], [data-group]'),
                $color = $item.find('.color'),
                $overlay;

            $overlay = overlayCreate('prompt', {
                message: $.t('graph.labl_color'),
                callbacks: {
                    validate: function (data) {
                        var value;

                        PANE_UNLOAD_LOCK = true;

                        value = adminGraphGetValue($item);

                        if (!data) {
                            $color
                                .addClass('auto')
                                .removeAttr('style');

                            if (value.options)
                                delete value.options.color;

                            return;
                        }

                        $color
                            .removeClass('auto')
                            .css('color', data);

                        value.options = $.extend(value.options || {}, {
                            color: data
                        });
                    }
                },
                labels: {
                    reset: {
                        text: $.t('main.labl_reset_default')
                    },
                    validate: {
                        text: $.t('graph.labl_color_set')
                    }
                },
                reset: ''
            });

            $overlay.find('input[name=value]')
                .attr('type', 'color')
                .val(!$color.hasClass('auto') ? rgbToHex($color.css('color')) : '#ffffff');
        });

        linkRegister('set-scale', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-serie], [data-group]'),
                $scale = $item.find('a[href=#set-scale]'),
                value = adminGraphGetValue($item);

            $.ajax({
                    url: urlPrefix + '/resources',
                    type: 'GET'
                }).pipe(function (data) {
                    var $input,
                        $overlay;

                    $overlay = overlayCreate('select', {
                        message: $.t('graph.labl_scale'),
                        value: value.scale,
                        callbacks: {
                            validate: function (data) {
                                data = parseFloat(data);
                                value.scale = data;
                                $scale.text(data ? data.toPrecision(3) : '');
                            }
                        },
                        labels: {
                            validate: {
                                text: $.t('graph.labl_scale_set')
                            }
                        },
                        reset: 0,
                        options: data.scales
                    });

                    $input = $overlay.find('input[name=value]');

                    $overlay.find('select')
                        .on('change', function (e) {
                            if (e.target.value)
                                $input.val(e.target.value);
                        })
                        .val(value.scale)
                        .trigger({
                            type: 'change',
                            _init: true
                        });
                });
        });

        // Attach events
        $body
            .on('click', 'button', function (e) {
                var $entry,
                    $entryActive,
                    $fieldset,
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
                        $entryActive = listGetItems($list, '.active');

                    name = $entryActive && $entryActive.data('value').name || null;

                    $entry = adminGraphCreateSerie(name, {
                        name: name,
                        origin: $origin.val(),
                        source: ($source.data('value') && $source.data('value').source.endsWith('groups') ?
                            'group:' : '') + $source.val(),
                        metric: ($metric.data('value') && $metric.data('value').source.endsWith('groups') ?
                            'group:' : '') + $metric.val()
                    });

                    if ($entryActive)
                        $entryActive.replaceWith($entry);

                    listSay($list, null);
                    listUpdateCount($list);

                    $origin.data('value', null).val('');
                    $source.data('value', null).val('');
                    $metric.data('value', null).val('');

                    $origin
                        .trigger('change')
                        .focus();

                    PANE_UNLOAD_LOCK = true;

                    break;

                case 'metric-cancel':
                    listMatch('step-1-metrics').find('[data-listitem^=step-1-metrics-item].active').trigger('click');

                    $(e.target).closest('fieldset').find('input[name=origin]')
                        .focus();

                    PANE_UNLOAD_LOCK = true;

                    break;

                case 'stack-config':
                    paneGoto('graph-edit', 'stack');
                    break;

                case 'step-cancel':
                    window.location = urlPrefix + '/admin/graphs/';
                    break;

                case 'step-save':
                    adminItemHandlePaneSave($(e.target).closest('[data-pane]'), graphId, 'graph', adminGraphGetData);
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
                    fieldValue,
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

                $fieldset.find('input[name=origin]')
                    .data('value', {
                        name: value.origin,
                        source: 'catalog/origins'
                    })
                    .val(active ? value.origin : '');

                if (value.source.startsWith('group:'))
                    fieldValue = {name: value.source.substr(6), source: 'library/sourcegroups'};
                else
                    fieldValue = {name: value.source, source: 'catalog/sources'};

                $fieldset.find('input[name=source]')
                    .data('value', fieldValue)
                    .val(active ? fieldValue.name : '');

                if (value.metric.startsWith('group:'))
                    fieldValue = {name: value.metric.substr(6), metric: 'library/metricgroups'};
                else
                    fieldValue = {name: value.metric, metric: 'catalog/metrics'};

                $fieldset.find('input[name=metric]')
                    .data('value', fieldValue)
                    .val(active ? fieldValue.name : '')
                    .trigger('change');
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
                if (!e._typing && $target.val()) {
                    $next = $target.closest('[data-input]').nextAll('[data-input], button:visible').first();

                    if ($next.attr('data-input') !== undefined)
                        $next = $next.children('input');

                    $next.focus();
                }
            })
            .on('change', '[data-step=3] select', function (e) {
                var $target = $(e.target);

                if (e._init || !e._select)
                    return;

                if (e.target.name == 'stack-mode') {
                    $target.closest('[data-step]').find('button[name=stack-config]')
                        .toggle(parseInt(e.target.value, 10) !== STACK_MODE_NONE);

                    paneGoto('graph-edit', 'stack', true);
                }

                itemSave(null, 'graphs', adminGraphGetData(), SAVE_MODE_VOLATILE)
                    .pipe(function (data, status, xhr) { /*jshint unused: true */
                        var location = xhr.getResponseHeader('Location');

                        graphDraw($target.closest('[data-step]').find('[data-graph]').attr('data-graph',
                            location.substr(location.lastIndexOf('/') + 1)));
                    });
            })
            .on('change', '[data-step=3] :input', function (e) {
                if (e._init || !e._select)
                    return;

                PANE_UNLOAD_LOCK = true;
            })
            .on('keydown', '[data-step=1] fieldset input', function (e) {
                $(e.target).nextAll('input')
                    .attr('disabled', 'disabled')
                    .val('');
            })
            .on('keyup', '[data-step=1] fieldset input', adminHandleFieldType)
            .on('dragstart dragend dragover dragleave drop', '.dragarea', adminGraphHandleSerieDrag);

        // Load graph data
        if (graphId === null)
            return;

        itemLoad(graphId, 'graphs').pipe(function (data) {
            var $itemOper,
                $itemSerie,
                $itemStack,
                $listMetrics,
                $listOpers,
                $listSeries,
                $pane,
                i,
                j,
                k;

            $listMetrics = listMatch('step-1-metrics');
            $listOpers   = listMatch('step-2-groups');
            $listSeries  = listMatch('step-stack-series');

            for (i in data.stacks) {
                $itemStack = data.stacks[i].mode !== STACK_MODE_NONE ? adminGraphCreateStack({
                    name: data.stacks[i].name
                }) : null;

                for (j in data.stacks[i].groups) {
                    $itemOper = data.stacks[i].groups[j].type !== OPER_GROUP_TYPE_NONE ? adminGraphCreateGroup(null, {
                        name: data.stacks[i].groups[j].name,
                        type: data.stacks[i].groups[j].type,
                        options: data.stacks[i].groups[j].options
                    }) : null;

                    for (k in data.stacks[i].groups[j].series) {
                        $itemSerie = adminGraphCreateSerie(null, $.extend(data.stacks[i].groups[j].series[k], {
                            options: data.stacks[i].groups[j].type === OPER_GROUP_TYPE_NONE ?
                                data.stacks[i].groups[j].options : null
                        }));

                        if ($itemOper)
                            adminGraphCreateProxy(PROXY_TYPE_SERIE, $itemSerie, $itemOper);
                        else if ($itemStack)
                            adminGraphCreateProxy(PROXY_TYPE_SERIE, $itemSerie, $itemStack);
                    }

                    if ($itemOper && $itemStack)
                        adminGraphCreateProxy(PROXY_TYPE_GROUP, $itemOper, $itemStack);
                }
            }

            $pane = paneMatch('graph-edit');

            $pane.find('input[name=graph-name]').val(data.name);
            $pane.find('textarea[name=graph-desc]').val(data.description);

            $pane.find('select[name=graph-type]').val(data.type).trigger({
                type: 'change',
                _init: true
            });

            $pane.find('select[name=stack-mode]').val(data.stack_mode).trigger({
                type: 'change',
                _init: true
            });

            if ($listMetrics.data('counter') === 0)
                listSay($listMetrics, $.t('metric.mesg_none'));

            if ($listSeries.data('counter') === 0)
                listSay($listSeries, $.t('graph.mesg_no_serie'));

            listUpdateCount($listMetrics);
        });
    });
}
