
function adminGraphGetData(link) {
    var $pane,
        data;

    link = typeof link == 'boolean' ? link : false;

    if (link) {
        $pane = paneMatch('graph-link-edit');

        // Create linked graph structure
        data = {
            name: $pane.find('input[name=graph-name]').val(),
            link: $pane.find('input[name=graph]').data('value').id,
            attributes: {}
        };

        $pane.find('.graphattrs [data-listitem]').each(function () {
            var $item = $(this);
            data.attributes[$item.find('.key :input').val()] = $item.find('.value :input').val();
        });
    } else {
        $pane = paneMatch('graph-edit');

        // Create standard graph structure
        data = {
            name: $pane.find('input[name=graph-name]').val(),
            description: $pane.find('textarea[name=graph-desc]').val(),
            title: $pane.find('input[name=graph-title]').val(),
            type: parseInt($pane.find('select[name=graph-type]').val(), 10),
            stack_mode: parseInt($pane.find('select[name=stack-mode]').val(), 10),
            unit_legend: $pane.find('input[name=unit-legend]').val(),
            unit_type: parseInt($pane.find('input[name=unit-type]:checked').val(), 10),
            groups: adminGraphGetGroups()
        };

        // Append graph arguments if template
        if ($pane.data('template')) {
            data.template = true;

            // Set extra pane redirection parameters
            $pane.data('redirect-params', 'templates=1');
        }
    }

    return data;
}

function adminGraphGetGroup(entry) {
    var group,
        value;

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
        value = $.extend({}, adminGraphGetValue(entry));

        group = {
            name: value && value.name || entry.attr('data-series'),
            type: OPER_GROUP_TYPE_NONE,
            series: [],
            options: $.extend({}, value.options)
        };

        delete value.options;
        group.series.push(value);
    }

    return group;
}

function adminGraphGetGroups() {
    var $listSeries = listMatch('step-stack-series'),
        $listStacks = listMatch('step-stack-groups'),
        count = 0,
        groups = [];

    // Retrieve defined stacks
    listGetItems($listStacks).each(function () {
        var $item = $(this);

        $item.find('.groupentry').each(function () {
            groups.push($.extend(adminGraphGetGroup($(this)), {stack_id: count}));
        });

        count++;
    });

    // Create new stack with remaining items
    listGetItems($listSeries, ':not(.linked)').each(function () {
        groups.push($.extend(adminGraphGetGroup($(this)), {stack_id: count}));
    });

    return groups;
}

function adminGraphGetValue(item) {
    if (item.data('source')) {
        if (item.hasClass('expand'))
            return item.data('source').data('expands')[item.attr('data-series')];
        else
            return item.data('source').data('value');
    } else {
        return item.data('value');
    }
}

function adminGraphCreateSeries(name, value) {
    var $item,
        $list = listMatch('step-1-metrics');

    // Set defaults
    if (!name)
        name = listNextName($list, 'data-series');

    if (!value.name)
        value.name = name;

    // Create new series
    $item = listAppend($list)
        .attr('data-series', name)
        .data({
            expands: {},
            proxies: [],
            renamed: false,
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

    if (value.type == OPER_GROUP_TYPE_AVERAGE)
        type = 'average';
    else if (value.type == OPER_GROUP_TYPE_SUM)
        type = 'sum';
    else if (value.type == OPER_GROUP_TYPE_NORMALIZE)
        type = 'normalize';
    else
        type = '';

    // Update group
    domFillItem($item, {
        name: value.name,
        type: type
    });

    if (value.options) {
        if (value.options.color)
            $item.find('.color')
                .removeClass('auto')
                .css('color', value.options.color);

        if (value.options.scale)
            $item.find('a[href=#set-scale]').text(value.options.scale.toPrecision(3));

        if (value.options.unit)
            $item.find('a[href=#set-unit]').text(value.options.unit);

        $item.find('a[href=#set-consolidate]').text(adminGraphGetConsolidateLabel(value.options.consolidate ||
            CONSOLIDATE_AVERAGE));

        if (value.options.formatter)
            $item.find('a[href=#set-formatter]').text(value.options.formatter);
    }

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
    case PROXY_TYPE_SERIES:
        attr = 'data-series';
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
        if (value.type == OPER_GROUP_TYPE_AVERAGE)
            value.type = 'avg';
        else if (value.type == OPER_GROUP_TYPE_SUM)
            value.type = 'sum';
        else
            delete value.type;
    }

    domFillItem($item, value);

    if (value.options) {
        if (value.options.color)
            $item.find('.color')
                .removeClass('auto')
                .css('color', value.options.color);

        if (value.options.scale)
            $item.find('a[href=#set-scale]').text(value.options.scale.toPrecision(3));

        if (value.options.unit)
            $item.find('a[href=#set-unit]').text(value.options.unit);

        $item.find('a[href=#set-consolidate]').text(adminGraphGetConsolidateLabel(value.options.consolidate ||
            CONSOLIDATE_AVERAGE));

        if (value.options.formatter)
            $item.find('a[href=#set-formatter]').text(value.options.formatter);
    }

    return $item;
}

function adminGraphHandleSeriesDrag(e) {
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
        if ($target.hasClass('linked') || !$target.attr('data-series') && !$target.attr('data-group'))
            return;

        $target.addClass('dragged');

        if ($target.attr('data-group'))
            e.dataTransfer.setData('text/plain', 'data-group=' + $target.attr('data-group'));
        else
            e.dataTransfer.setData('text/plain', 'data-series=' + $target.attr('data-series'));

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
            listSay($listSeries, $.t('graph.mesg_no_series'));

        // Handle drop'n'create
        if (e.target.tagName == 'A') {
            $target.trigger('click');
            $target = listMatch('step-' + ADMIN_PANES['graph-edit'].active + '-groups').find('.groupitem:last');
        }

        chunks = e.dataTransfer.getData('text/plain').split('=');

        if (chunks[0] == 'data-series') {
            $itemSrc = listMatch('step-2-series').find('[' + e.dataTransfer.getData('text/plain') + ']');

            $item = adminGraphCreateProxy(PROXY_TYPE_SERIES, $itemSrc.data('source'), $target);

            if ($itemSrc.hasClass('expand')) {
                $item.attr('data-series', chunks[1]);
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

function adminGraphAutoNameSeries(force) {
    var $items = listGetItems('step-1-metrics'),
        refCounts = {
            origin: [],
            source: [],
            metric: {}
        },
        refs = [],
        prefix,
        prefixLen,
        i;

    force = typeof force == 'boolean' ? force : false;

    $items.each(function () {
        var $item = $(this),
            value = adminGraphGetValue($item),
            fullName = value.origin+'/'+value.source+'/'+value.metric;

        if (refCounts.origin.indexOf(value.origin) == -1)
            refCounts.origin.push(value.origin);

        if (refCounts.source.indexOf(value.source) == -1)
            refCounts.source.push(value.source);

        if (!refCounts.metric[fullName]) {
            refCounts.metric[fullName] = {current: 1, count: 1};
        } else {
            refCounts.metric[fullName].current++;
            refCounts.metric[fullName].count++;
        }
    });

    $items.each(function (i) {
        var $item = $(this),
            value,
            fullName,
            matchLen;

        value = adminGraphGetValue($item);
        fullName = value.origin+'/'+value.source+'/'+value.metric;

        value.name = value.metric;

        if (refCounts.origin.length > 1) {
            value.name = value.origin + '/' + value.source + '/' + value.name;
        } else if (refCounts.source.length > 1) {
            value.name = value.source + '/' + value.name;
        }

        if (refCounts.metric[fullName].count > 1) {
            value.name += ' (' + (refCounts.metric[fullName].count - refCounts.metric[fullName].current) + ')';
            refCounts.metric[fullName].current--;
        }

        if (i === 0) {
            prefix = value.name;
            prefixLen = prefix.length;
        } else {
            matchLen = 0;

            while (++matchLen < prefixLen && matchLen < value.name.length) {
                if (value.name.charAt(matchLen) != prefix.charAt(matchLen))
                    break;
            }

            prefixLen = matchLen;
        }

        refs.push([$item, value]);
    });

    // Substract trailing word characters
    if (prefix && prefixLen)
        prefixLen -= (prefix.substr(0, prefixLen).match(/\w+$/) || '').length;

    for (i in refs) {
        if (refs[i][0].data('renamed') && !force)
            return;
        else if (force)
            refs[i][0].data('renamed', false);

        if (prefixLen > 1 && refs.length > 1)
            refs[i][1].name = refs[i][1].name.substr(prefixLen);

        domFillItem(refs[i][0], refs[i][1]);
    }
}

function adminGraphGetConsolidateLabel(type) {
    switch (type) {
        case CONSOLIDATE_AVERAGE:
            return 'avg';
        case CONSOLIDATE_LAST:
            return 'last';
        case CONSOLIDATE_MAX:
            return 'max';
        case CONSOLIDATE_MIN:
            return 'min';
        case CONSOLIDATE_SUM:
            return 'sum';
        default:
            return '';
    }
}

function adminGraphGetTemplatable(groups) {
    var $pane = paneMatch('graph-edit'),
        result = [],
        regexp,
        series,
        i, j;

    regexp = /\{\{\s*\.([a-z0-9]+)\s*\}\}/i;

    for (i in groups) {
        for (j in groups[i].series) {
            series = groups[i].series[j];
            result = result.concat((series.origin + '\x1e' + series.source + '\x1e' + series.metric).matchAll(regexp));
        }
    }

    if ($pane.data('template')) {
        result = result.concat($pane.find('textarea[name=graph-desc]').val().matchAll(regexp));
        result = result.concat($pane.find('input[name=graph-title]').val().matchAll(regexp));
    }

    result.sort();

    return arrayUnique(result);
}

function adminGraphUpdateAttrsList() {
    var $pane = paneMatch('graph-edit'),
        $listAttrs,
        $item,
        attrs,
        attrsData,
        i;

    // Generate graph arguments list
    $listAttrs = listMatch('step-3-attrs');

    attrs = adminGraphGetTemplatable(adminGraphGetGroups());

    if (attrs.length === 0) {
        listSay($listAttrs, $.t('graph.mesg_no_template_attr'), 'warning');
        $listAttrs.next('.mesgitem').hide();
        return;
    } else {
        $listAttrs.next('.mesgitem').show();
    }

    listSay($listAttrs, null);
    listEmpty($listAttrs);

    attrsData = $pane.data('attrs-data') || {};

    for (i in attrs) {
        $item = listAppend($listAttrs);
        $item.find('.key input').val(attrs[i]);

        if (attrsData[attrs[i]] !== undefined)
            $item.find('.value input').val(attrsData[attrs[i]]);
    }
}

function adminGraphSetupTerminate() {
    // Register admin panes
    paneRegister('graph-list', function () {
        listRegisterItemCallback('graphs', function (item, entry) {
            if (entry.template)
                item.find('a[href=#show-graph]').remove();
            else
                item.find('a[href=#add-graph]').remove();

            if (!entry.link)
                return;

            item.data('params', 'linked=1');

            item.find('.name')
                .attr('title', $.t('graph.mesg_linked'))
                .addClass('linked');
        });

        adminItemHandlePaneList('graph');

        // Register links
        linkRegister('add-graph', function (e) {
            window.location = urlPrefix + '/admin/graphs/add?linked=1&from=' +
                $(e.target).closest('[data-itemid]').attr('data-itemid');
        });
    });

    paneRegister('graph-edit', function () {
        var $pane = paneMatch('graph-edit'),
            graphId = $pane.opts('pane').id || null;

        // Register completes and checks
        if ($('[data-input=source]').length > 0) {
            inputRegisterComplete('source', function (input) {
                var $origin = input.closest('fieldset').find('input[name=origin]'),
                    params = {},
                    opts;

                opts = $origin.closest('[data-input]').opts('input');
                if (!opts.ignorepattern || $origin.val().indexOf(opts.ignorepattern) == -1)
                    params.origin = $origin.val();

                return inputGetSources(input, params);
            });
        }

        if ($('[data-input=metric]').length > 0) {
            inputRegisterComplete('metric', function (input) {
                var $fieldset = input.closest('fieldset'),
                    $origin = $fieldset.find('input[name=origin]'),
                    $source = $fieldset.find('input[name=source]'),
                    params = {},
                    opts;

                opts = $origin.closest('[data-input]').opts('input');
                if (!opts.ignorepattern || $origin.val().indexOf(opts.ignorepattern) == -1)
                    params.origin = $origin.val();

                opts = $source.closest('[data-input]').opts('input');
                if (!opts.ignorepattern || $source.val().indexOf(opts.ignorepattern) == -1)
                    params.source = ($source.data('value') && $source.data('value').source.endsWith('groups/') ?
                        'group:' : '') + $source.val();

                return inputGetSources(input, params);
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
                    if (data.length > 0 && data[0].id != graphId) {
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

            $pane.find('button[name=auto-name]').show();

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

            $pane.find('button[name=auto-name]').hide();

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

                $item = adminGraphCreateProxy(PROXY_TYPE_SERIES, $itemSrc, $listSeries);

                if (value.source.startsWith('group:') || value.metric.startsWith('group:')) {
                    expandQuery.push([value.origin, value.source, value.metric]);
                    expand = true;
                    $item.addClass('expandable');
                }

                if ($listOpers.find('[data-series="' + $itemSrc.attr('data-series') + '"]').length > 0)
                    $item.addClass('linked');
            });

            adminGraphRestorePosition();

            if (listGetCount($listSeries, ':not(.linked)') > 0)
                listSay($listSeries, null);

            // Retrieve expanding information
            if (expand) {
                $.ajax({
                    url: urlPrefix + '/api/v1/library/expand',
                    type: 'POST',
                    contentType: 'application/json',
                    data: JSON.stringify(expandQuery)
                }).pipe(function (data) {
                    if (!data)
                        return;

                    listGetItems($listSeries).each(function (index) {
                        var $item = $(this);

                        if (!$item.hasClass('expandable')) {
                            $item.find('.count').remove();
                            $item.find('a[href$=#expand-series], a[href$=#collapse-series]').remove();
                            return;
                        }

                        if (data[index] && data[index].length > 0) {
                            $item.find('.count').text(data[index].length);

                            if (data[index].length > 1) {
                                $item.data('expand', data[index]);
                                $item.find('a[href$=#collapse-series]').remove();
                            }
                        } else {
                            $item.find('.count').text(0);
                            $item.find('a[href$=#expand-series], a[href$=#collapse-series]').remove();
                        }

                        // Restore expanded state
                        if (!$.isEmptyObject(listMatch('step-1-metrics').find('[data-series="' +
                                $item.attr('data-series') + '"]').data('expands')))
                            $item.find('a[href$=#expand-series]').trigger('click');
                    });
                });
            } else {
                $listSeries.find('.count').remove();
                $listSeries.find('a[href$=#expand-series], a[href$=#collapse-series]').remove();
            }
        });

        paneStepRegister('graph-edit', 3, function () {
            var $step = $('[data-step=3]');

            $pane.find('button[name=auto-name]').hide();

            $step.find('select:last').trigger('change');

            setTimeout(function () {
                selectUpdate($step.find('select[name=graph-unit]').get(0));
                $step.find(':input:first').select();
            });

            // Check for template
            $pane.find('.tmplattrs').toggle($pane.data('template'));

            if (!$pane.data('template'))
                return;

            // Generate graph arguments list
            adminGraphUpdateAttrsList();
        });

        paneStepRegister('graph-edit', 'stack', function () {
            var $listSeries = listMatch('step-stack-series'),
                $listStacks = listMatch('step-stack-groups');

            $pane.find('button[name=auto-name]').hide();

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
                    adminGraphCreateProxy(PROXY_TYPE_SERIES, $(this).data('source'), $item);
                });
            });

            // Create groups for each remaining items
            listGetItems('step-2-series').each(function () {
                var $item,
                    $itemMain,
                    $itemSrc = $(this),
                    value;

                if ($itemSrc.hasClass('linked')) {
                    $listStacks.find('[data-series="' + $itemSrc.attr('data-series') + '"]').remove();
                    return;
                }

                $itemMain = adminGraphCreateProxy(PROXY_TYPE_SERIES, $itemSrc.data('source'), $listSeries)
                    .attr('data-series', $itemSrc.attr('data-series'));

                if ($listStacks.find('[data-series="' + $itemSrc.attr('data-series') + '"]').length > 0)
                    $itemMain.addClass('linked');

                $item = adminGraphCreateProxy(PROXY_TYPE_SERIES, $itemSrc.data('source'), $itemMain);

                if ($itemSrc.hasClass('expand')) {
                    value = $itemSrc.data('source').data('expands')[$itemSrc.attr('data-series')];
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
        linkRegister('add-average add-sum add-normalize add-stack', function (e) {
            var operGroupType;

            if (e.target.href.endsWith('-stack')) {
                // Add stack group
                adminGraphCreateStack({});
            } else {
                if (e.target.href.endsWith('-average'))
                    operGroupType = OPER_GROUP_TYPE_AVERAGE;
                else if (e.target.href.endsWith('-sum'))
                    operGroupType = OPER_GROUP_TYPE_SUM;
                else if (e.target.href.endsWith('-normalize'))
                    operGroupType = OPER_GROUP_TYPE_NORMALIZE;
                else
                    return;

                adminGraphCreateGroup(null, {type: operGroupType});
            }

            PANE_UNLOAD_LOCK = true;
        });

        linkRegister('collapse-series', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-series]'),
                $series,
                collapse,
                name = $item.attr('data-series').split('-')[0];

            // Collapse expanded series
            $series = listMatch('step-2-groups').find('[data-series^="' + name + '-"]');

            collapse = function () {
                $item.data('source').data('expands', {});

                $item.siblings('[data-series="' + name + '"]').removeClass('linked');
                $item.siblings('[data-series^="' + name + '-"]').andSelf().remove();

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

        linkRegister('expand-series', function (e) {
            var $target = $(e.target),
                $item,
                $itemSrc = $target.closest('[data-series]'),
                $itemRef = $itemSrc,
                $listSeries = $itemSrc.closest('[data-list]'),
                data = $itemSrc.data('expand'),
                expands = $itemSrc.data('source').data('expands'),
                i,
                name = $itemSrc.attr('data-series'),
                seriesName;

            // Expand series
            for (i in data) {
                seriesName = name + '-' + i;

                $item = adminGraphCreateProxy(PROXY_TYPE_SERIES, $itemSrc.data('source'), $listSeries)
                    .attr('data-series', seriesName)
                    .addClass('expand');

                $item.detach().insertAfter($itemRef);

                if (!expands[seriesName]) {
                    expands[seriesName] = {
                        name: (adminGraphGetValue($itemSrc).name || name) + ' (' + i + ')',
                        origin: data[i][0],
                        source: data[i][1],
                        metric: data[i][2]
                    };
                }

                domFillItem($item, expands[seriesName]);

                if (expands[seriesName].options) {
                    if (expands[seriesName].options.color)
                        $item.find('.color')
                            .removeClass('auto')
                            .css('color', expands[seriesName].options.color);

                    if (expands[seriesName].options.scale !== 0)
                        $item.find('a[href=#set-scale]').text(expands[seriesName].options.scale);

                    if (expands[seriesName].options.unit !== 0)
                        $item.find('a[href=#set-unit]').text(expands[seriesName].options.unit);

                    if (expands[seriesName].options.consolidate !== 0)
                        $item.find('a[href=#set-consolidate]').text(expands[seriesName].options.consolidate);

                    if (expands[seriesName].options.formatter !== 0)
                        $item.find('a[href=#set-formatter]').text(expands[seriesName].options.formatter);
                }

                $item.find('.count').remove();
                $item.find('a[href$=#expand-series]').remove();

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
                $itemNext = $item.prevAll('.listitem, .groupitem, .groupentry').filter(':not(.linked):first');

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
                $item = $target.closest('[data-step]').find('[data-series="' + $entry.attr('data-series') + '"]')
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

            // Unselect active item before removal
            if ($item.hasClass('active'))
                $item.trigger('click');

            // Remove proxy items
            data = $item.data('proxies');

            for (i in data)
                data[i].remove();

            $item.remove();

            listUpdateCount($list);

            if (listGetCount($list) === 0)
                listSay($list, $.t('metric.mesg_none'), 'info');

            adminGraphAutoNameSeries();

            PANE_UNLOAD_LOCK = true;

            $('[data-step=1] fieldset input[name=origin]').focus();
        });

        linkRegister('rename-series rename-group rename-stack', function (e) {
            var $target = $(e.target),
                $input,
                $item,
                $overlay,
                attrName,
                seriesName,
                value;

            if (e.target.href.endsWith('#rename-stack'))
                attrName = 'data-stack';
            else if (e.target.href.endsWith('#rename-group'))
                attrName = 'data-group';
            else
                attrName = 'data-series';

            $item     = $target.closest('[' + attrName + ']');
            seriesName = $item.attr(attrName);

            value = adminGraphGetValue($item).name;

            $overlay = overlayCreate('prompt', {
                message: $.t(attrName == 'data-stack' ? 'graph.labl_stack_name' : 'graph.labl_series_name'),
                callbacks: {
                    validate: function (data) {
                        if (!data)
                            return;

                        adminGraphGetValue($item).name = data;

                        paneMatch('graph-edit').find('[' + attrName + '="' + seriesName + '"]').each(function () {
                            $(this).find('.name:first').text(data);
                        });

                        if (attrName == 'data-series') {
                            $item.data('source').data('renamed', true);
                            $pane.find('button[name=auto-name]').removeAttr('disabled');
                        }

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
                        seriesName;

                    values.push($(this).data('value').name);

                    if (!$item.data('expands'))
                        return;

                    for (seriesName in $item.data('expands'))
                        values.push($item.data('expands')[seriesName].name);
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
                $item = $target.closest('[data-series], [data-group]'),
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
                $item = $target.closest('[data-series], [data-group]'),
                $scale = $item.find('a[href=#set-scale]'),
                value = adminGraphGetValue($item);

            $.ajax({
                url: urlPrefix + '/api/v1/library/scales/values',
                type: 'GET'
            }).pipe(function (data) {
                var $input,
                    $overlay,
                    options = [],
                    scaleValue = value.options && value.options.scale ? value.options.scale : '';

                $.each(data, function (i, entry) { /*jshint unused: true */
                    options.push([entry.name, entry.value]);
                });

                $overlay = overlayCreate('select', {
                    message: $.t('graph.labl_scale'),
                    value: scaleValue,
                    callbacks: {
                        validate: function (data) {
                            data = parseFloat(data);

                            value.options = $.extend(value.options || {}, {
                                scale: data
                            });

                            $scale.text(data ? data.toPrecision(3) : '');
                        }
                    },
                    labels: {
                        validate: {
                            text: $.t('graph.labl_scale_set')
                        }
                    },
                    reset: 0,
                    options: options
                });

                $input = $overlay.find('input[name=value]');

                $overlay.find('select')
                    .on('change', function (e) {
                        if (e.target.value)
                            $input.val(e.target.value);
                    })
                    .val(scaleValue)
                    .trigger({
                        type: 'change',
                        _init: true
                    });
            });
        });

        linkRegister('set-unit', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-series], [data-group]'),
                $unit = $item.find('a[href=#set-unit]'),
                value = adminGraphGetValue($item);

            $.ajax({
                url: urlPrefix + '/api/v1/library/units/labels',
                type: 'GET'
            }).pipe(function (data) {
                var $input,
                    $overlay,
                    options = [],
                    unitValue = value.options && value.options.unit ? value.options.unit : '';

                $.each(data, function (i, entry) { /*jshint unused: true */
                    options.push([entry.name, entry.label]);
                });

                $overlay = overlayCreate('select', {
                    message: $.t('graph.labl_unit'),
                    value: unitValue,
                    callbacks: {
                        validate: function (data) {
                            value.options = $.extend(value.options || {}, {
                                unit: data
                            });

                            $unit.text(data || '');
                        }
                    },
                    labels: {
                        validate: {
                            text: $.t('graph.labl_unit_set')
                        }
                    },
                    reset: 0,
                    options: options
                });

                $input = $overlay.find('input[name=value]');

                $overlay.find('select')
                    .on('change', function (e) {
                        if (e.target.value)
                            $input.val(e.target.value);
                    })
                    .val(unitValue)
                    .trigger({
                        type: 'change',
                        _init: true
                    });
            });
        });

        linkRegister('set-consolidate', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-series], [data-group]'),
                $input,
                $overlay,
                value = adminGraphGetValue($item),
                consolidateValue = value.options && value.options.consolidate ?
                    value.options.consolidate : CONSOLIDATE_AVERAGE;

            $overlay = overlayCreate('select', {
                message: $.t('graph.labl_consolidate'),
                value: consolidateValue,
                callbacks: {
                    validate: function (data) {
                        data = parseInt(data, 10);

                        value.options = $.extend(value.options || {}, {
                            consolidate: data
                        });

                        $item.find('a[href=#set-consolidate]').text(adminGraphGetConsolidateLabel(data));
                    }
                },
                labels: {
                    validate: {
                        text: $.t('graph.labl_consolidate_set')
                    }
                },
                reset: 0,
                options: [
                    [$.t('graph.labl_consolidate_average'), CONSOLIDATE_AVERAGE],
                    [$.t('graph.labl_consolidate_last'), CONSOLIDATE_LAST],
                    [$.t('graph.labl_consolidate_max'), CONSOLIDATE_MAX],
                    [$.t('graph.labl_consolidate_min'), CONSOLIDATE_MIN],
                    [$.t('graph.labl_consolidate_sum'), CONSOLIDATE_SUM],
                ]
            });

            $overlay.find('button[name=reset]').hide();

            $overlay.find('.select')
               .addClass('full')
               .find('.menu .menuitem:first').remove();

            $input = $overlay.find('input[name=value]').hide();

            $overlay.find('select')
                .on('change', function (e) {
                    if (e.target.value)
                        $input.val(e.target.value);
                })
                .val(consolidateValue)
                .trigger({
                    type: 'change',
                    _init: true
                });
        });

        linkRegister('set-formatter', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-series], [data-group]'),
                $formatter = $item.find('a[href=#set-formatter]'),
                value = adminGraphGetValue($item);

            overlayCreate('prompt', {
                message: $.t('graph.labl_formatter'),
                callbacks: {
                    validate: function (data) {
                        value.options = $.extend(value.options || {}, {
                            formatter: data
                        });

                        $formatter.text(data || '');
                    }
                },
                labels: {
                    validate: {
                        text: $.t('graph.labl_formatter_set')
                    }
                },
                reset: ''
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
                    name,
                    metricName;

                // Find closest button for browsers triggering event from children element
                if (e.target.tagName != 'BUTTON')
                    e.target = $(e.target).closest('button').get(0);

                switch (e.target.name) {
                case 'auto-name':
                    if (e.target.disabled)
                        return;

                    adminGraphAutoNameSeries(true);
                    $pane.find('button[name=auto-name]').attr('disabled', 'disabled');

                    break;

                case 'metric-add':
                case 'metric-update':
                    if (e.target.disabled)
                        return;

                    $list = listMatch('step-1-metrics');

                    $fieldset = $(e.target).closest('fieldset');
                    $metric   = $fieldset.find('input[name=metric]');
                    $source   = $fieldset.find('input[name=source]');
                    $origin   = $fieldset.find('input[name=origin]');

                    // Set template mode
                    if ($origin.val().indexOf('{{') != -1 || $source.val().indexOf('{{') != -1 ||
                        $metric.val().indexOf('{{') != -1) {
                        $pane.data('template', true);
                        $pane.find('button[name=step-save]').children().hide().filter('.template').show();
                    } else if (e.target.name == 'metric-update' && listGetCount($list) == 1) {
                        $pane.data('template', false);
                        $pane.find('button[name=step-save]').children().hide().filter('.default').show();
                    }

                    if (e.target.name == 'metric-update')
                        $entryActive = listGetItems($list, '.active');

                    metricName = ($metric.data('value') && $metric.data('value').source.endsWith('groups/') ?
                        'group:' : '') + $metric.val();

                    name = $entryActive && $entryActive.attr('data-series') || null;

                    $entry = adminGraphCreateSeries(name, {
                        name: name || metricName,
                        origin: $origin.val(),
                        source: ($source.data('value') && $source.data('value').source.endsWith('groups/') ?
                            'group:' : '') + $source.val(),
                        metric: metricName
                    });

                    if ($entryActive) {
                        $entry.insertBefore($entryActive);
                        $entryActive.find('a[href=#remove-metric]').trigger('click');
                    }

                    listSay($list, null);
                    listUpdateCount($list);

                    $metric.data('value', null).val('');

                    $metric
                        .trigger('change')
                        .focus();

                    $fieldset.find('button[name=metric-add]').show();
                    $fieldset.find('button[name=metric-update], button[name=metric-cancel]').hide();

                    adminGraphAutoNameSeries();

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
                    fieldValue = {name: value.source.substr(6), source: 'library/sourcegroups/'};
                else
                    fieldValue = {name: value.source, source: 'catalog/sources/'};

                $fieldset.find('input[name=source]')
                    .data('value', fieldValue)
                    .val(active ? fieldValue.name : '');

                if (value.metric.startsWith('group:'))
                    fieldValue = {name: value.metric.substr(6), source: 'library/metricgroups/'};
                else
                    fieldValue = {name: value.metric, source: 'catalog/metrics/'};

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

                    if (!e._autofill || $next.prop("tagName") != 'BUTTON')
                        $next.focus();
                }
            })
            .on('change', '[data-step=3] select, [data-step=3] input[type=radio]', function (e) {
                var $target = $(e.target),
                    data;

                if (e._init || !e._select && e.target.tagName == 'SELECT')
                    return;

                if (e.target.name == 'stack-mode') {
                    $target.closest('[data-step]').find('button[name=stack-config]')
                        .toggle(parseInt(e.target.value, 10) !== STACK_MODE_NONE);

                    paneGoto('graph-edit', 'stack', true);
                }

                data = adminGraphGetData();

                if ($pane.data('template')) {
                    data.attributes = {};

                    $pane.find('.graphattrs [data-listitem]').each(function () {
                        var $item = $(this);
                        data.attributes[$item.find('.key :input').val()] = $item.find('.value :input').val();
                    });

                    // Save attributes data for pane-switch restoration
                    $pane.data('attrs-data', data.attributes);
                }

                graphDraw($target.closest('[data-step]').find('[data-graph]'), false, 0, data);
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
            .on('keypress', '[data-step=3] :input[name=graph-desc], [data-step=3] :input[name=graph-title]',
                function (e) {

                var $target = $(e.target),
                    $step = $target.closest('[data-step]');

                if ($step.data('attrs-timeout')) {
                    clearTimeout($step.data('attrs-timeout'));
                    $step.removeData('attrs-timeout');
                }

                $step.data('attrs-timeout', setTimeout(adminGraphUpdateAttrsList, 500));
            })
            .on('keypress', '[data-step=3] .graphattrs :input', function (e) {
                var $target,
                    $attrs;

                if (!$pane.data('template'))
                    return;

                $target = $(e.target);
                $attrs = $target.closest('.graphattrs');

                if ($attrs.data('timeout')) {
                    clearTimeout($attrs.data('timeout'));
                    $attrs.removeData('timeout');
                }

                // Trigger graph redraw
                $attrs.data('timeout', setTimeout(function () {
                    $pane.find('[data-step=3] select:first').trigger('change');
                }, 1000));
            })
            .on('dragstart dragend dragover dragleave drop', '.dragarea', adminGraphHandleSeriesDrag);

        // Set default panel save button
        $pane.find('button[name=step-save]').children().hide().filter('.default').show();

        // Load graph data
        if (graphId === null)
            return;

        itemLoad(graphId, 'graphs').pipe(function (data) {
            var $itemOper,
                $itemSeries,
                $listMetrics,
                $listOpers,
                $listSeries,
                stacks = {},
                i,
                j;

            $listMetrics = listMatch('step-1-metrics');
            $listOpers   = listMatch('step-2-groups');
            $listSeries  = listMatch('step-stack-series');

            for (i in data.groups) {
                if (!stacks[data.groups[i].stack_id]) {
                    stacks[data.groups[i].stack_id] = data.stack_mode !== STACK_MODE_NONE ? adminGraphCreateStack({
                        name: 'stack' + data.groups[i].stack_id
                    }) : null;
                }

                $itemOper = data.groups[i].type !== OPER_GROUP_TYPE_NONE ? adminGraphCreateGroup(null, {
                    name: data.groups[i].name,
                    type: data.groups[i].type,
                    options: data.groups[i].options
                }) : null;

                for (j in data.groups[i].series) {
                    if (data.groups[i].type === OPER_GROUP_TYPE_NONE && !data.groups[i].series[j].options)
                        data.groups[i].series[j].options = data.groups[i].options;


                    if (data.groups[i].series[j].origin.indexOf('{{') != -1 ||
                        data.groups[i].series[j].source.indexOf('{{') != -1 ||
                        data.groups[i].series[j].metric.indexOf('{{') != -1) {
                        $pane.data('template', true);
                    } else {
                        $pane.data('template', false);
                    }

                    $itemSeries = adminGraphCreateSeries(null, data.groups[i].series[j])
                        .data('renamed', true);

                    if ($itemOper)
                        adminGraphCreateProxy(PROXY_TYPE_SERIES, $itemSeries, $itemOper);
                    else if (stacks[data.groups[i].stack_id])
                        adminGraphCreateProxy(PROXY_TYPE_SERIES, $itemSeries, stacks[data.groups[i].stack_id]);
                }

                if ($itemOper && stacks[data.groups[i].stack_id])
                    adminGraphCreateProxy(PROXY_TYPE_GROUP, $itemOper, stacks[data.groups[i].stack_id]);
            }

            if ($pane.data('template'))
                $pane.find('button[name=step-save]').children().hide().filter('.template').show();

            $pane.find('input[name=graph-name]').val(data.name);
            $pane.find('textarea[name=graph-desc]').val(data.description);

            $pane.find('input[name=graph-title]').val(data.title);

            $pane.find('select[name=graph-type]').val(data.type).trigger({
                type: 'change',
                _init: true
            });

            $pane.find('input[name=unit-legend]').val(data.unit_legend);
            $pane.find('input[name=unit-type][value=' + data.unit_type + ']').prop('checked', true);

            $pane.find('select[name=stack-mode]').val(data.stack_mode).trigger({
                type: 'change',
                _init: true
            });

            if ($listMetrics.data('counter') === 0)
                listSay($listMetrics, $.t('metric.mesg_none'));

            if ($listSeries.data('counter') === 0)
                listSay($listSeries, $.t('graph.mesg_no_series'));

            listUpdateCount($listMetrics);
        });
    });

    paneRegister('graph-link-edit', function () {
        var $pane = paneMatch('graph-link-edit'),
            graphId = $pane.opts('pane').id || null;

        // Register checks
        if ($('[data-input=graph-name]').length > 0) {
            inputRegisterCheck('graph-name', function (input) {
                var value = input.find(':input').val();

                if (!value)
                    return;

                itemList({
                    filter: value
                }, 'graphs').pipe(function (data) {
                    if (data.length > 0 && data[0].id != graphId) {
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
        paneStepRegister('graph-link-edit', 1, function () {
            var linkSource;

            if (!graphId)
                listSay('step-1-attrs', $.t('graph.mesg_no_template_selected'), 'info');

            linkSource = getURLParams().from;
            if (linkSource) {
                itemLoad(linkSource, 'graphs').pipe(function (data) {
                    inputMatch('graph').find(':input')
                        .data('value', {
                            id: data.id,
                            name: data.name,
                            description: data.description,
                            modified: data.modified,
                            template: data.template,
                            source: 'library/graphs/?type=template'
                        })
                        .val(data.name)
                        .trigger('change');

                    $('button[name=graph-ok]').trigger('click');
                });
            }

            setTimeout(function () { $('[data-step=1] input').trigger('change').filter(':first').select(); }, 0);
        });

        // Register links
        linkRegister('edit-template', function () {
            window.location = urlPrefix + '/admin/graphs/' + $pane.find('input[name=graph]').data('value').id;
        });

        // Attach events
        $body
            .on('click', 'button', function (e) {
                var $graph,
                    $list,
                    $target = $(e.target);

                switch (e.target.name) {
                case 'graph-ok':
                    if (e.target.disabled)
                        return;

                    $list = listMatch('step-1-attrs');
                    $graph = $pane.find('input[name=graph]');

                    $.ajax({
                        url: urlPrefix + '/api/v1/library/graphs/' + $graph.data('value').id,
                        type: 'GET',
                        dataType: 'json'
                    }).pipe(function (data) {
                        var $item,
                            attrs = adminGraphGetTemplatable(data.groups),
                            attrsData = $pane.data('attrs-data'),
                            i;

                        // Restore field name if needed (useful for linked graph edition)
                        if (!$graph.val())
                            $graph.val(data.name);

                        if (attrs.length === 0) {
                            listSay($list, $.t('graph.mesg_no_template_attr'), 'error');
                            $pane.find('button[name=step-save]').attr('disabled', 'disabled');
                            return;
                        } else {
                            $pane.find('button[name=step-save]').removeAttr('disabled');
                        }

                        listSay($list, null);
                        listEmpty($list);

                        for (i in attrs) {
                            $item = listAppend($list);
                            $item.find('.key input').val(attrs[i]);

                            if (attrsData && attrsData[attrs[i]] !== undefined)
                                $item.find('.value input').val(attrsData[attrs[i]]);
                        }

                        // Trigger first graph preview
                        listGetItems($list, ':first').find('.value input').trigger('keypress');

                        if (e._init)
                            return;

                        listGetItems($list, ':first').find('.value input').focus();
                    });

                    PANE_UNLOAD_LOCK = true;

                    break;

                case 'step-cancel':
                    window.location = urlPrefix + '/admin/graphs/';
                    break;

                case 'step-save':
                    if (!$pane.find('input[name=graph]').data('value')) {
                        overlayCreate('alert', {
                            message: $.t('graph.mesg_missing_template'),
                            callbacks: {
                                validate: function () {
                                    setTimeout(function () { $pane.find('[data-input=graph] input').select(); }, 0);
                                }
                            }
                        });
                        return false;
                    }

                    adminItemHandlePaneSave($target.closest('[data-pane]'), graphId, 'graph', function() {
                        return adminGraphGetData(true);
                    });

                    break;
                }
            })
            .on('change', '[data-step=1] [data-input=graph] input', function (e) {
                var $target = $(e.target),
                    $button = $target.closest('[data-input]').nextAll('button:first');

                if (!$target.val())
                    $button.attr('disabled', 'disabled');
                else
                    $button.removeAttr('disabled');

                // Select button
                if ($target.val())
                    $button.focus();
            })
            .on('keypress', '[data-step=1] .graphattrs :input', function (e) {
                var $target = $(e.target);

                if ($target.data('timeout')) {
                    clearTimeout($target.data('timeout'));
                    $target.removeData('timeout');
                }

                $target.data('timeout', setTimeout(function () {
                    graphDraw($target.closest('[data-step]').find('[data-graph]'), false, 0, adminGraphGetData(true));
                }, 1000));
            });

        // Load graph data
        if (graphId === null)
            return;

        itemLoad(graphId, 'graphs').pipe(function (data) {
            var $graph;

            $pane.find('input[name=graph-name]').val(data.name).select();

            $pane.data('attrs-data', data.attributes);

            $graph = $pane.find('input[name=graph]').data('value', {id: data.link});

            $graph.closest('[data-input]').nextAll('button:first')
                .removeAttr('disabled')
                .trigger({
                    type: 'click',
                    _init: true
                });

            PANE_UNLOAD_LOCK = false;
        });
    });
}
