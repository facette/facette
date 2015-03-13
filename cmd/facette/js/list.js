
/* List */

var LIST_CALLBACKS = {},
    LIST_TIMEOUTS = {};

function listAppend(list, refNode) {
    var $item;

    if (typeof list == 'string')
        list = listMatch(list);

    // Append new item
    $item = list.data('template').clone()
        .attr('data-listitem', list.attr('data-list') + '-item' + list.data('counter'));

    list.data('counter', list.data('counter') + 1);

    if (refNode)
        $item.insertAfter(refNode);
    else
        $item.appendTo(list.data('container'));

    if ($item.is('[data-list]'))
        listInit($item.get(0));

    $item.find('[data-list]').each(function () {
        listInit(this);
    });

    return $item;
}

function listEmpty(list) {
    if (typeof list == 'string')
        list = listMatch(list);

    list.data({
        counter: 0,
        offset: 0
    });

    listGetItems(list).remove();

    listUpdateCount(list, 0);
}

function listGetCount(list, filter) {
    return listGetItems(list, filter).length;
}

function listGetItems(list, filter) {
    if (typeof list == 'string')
        list = listMatch(list);

    return list.find('[data-listitem^="' + list.attr('data-list') + '-item"]' + (filter || ''));
}

function listInit(element) {
    return $.Deferred(function ($deferred) {
        var $item = $(element),
            $template;

        if (!$.contains(document.documentElement, element)) {
            $deferred.resolve();
            return;
        }

        $template = $item.find('[data-listtmpl="' + element.getAttribute('data-list') + '"]')
            .removeAttr('data-listtmpl');

        $item.data({
            counter: 0,
            offset: 0,
            template: $template,
            container: $template.parent()
        });

        $template.detach();

        // Initialize list content
        listSay($item, null);

        if ($item.opts('list').init) {
            listUpdate($item).then(function () { $deferred.resolve(); });
        } else {
            listSay($item, $.t(($item.opts('list').messages || 'item') + '.mesg_none'), 'info');
            $deferred.resolve();
        }
    }).promise();
}

function listMatch(name, suffix) {
    suffix = suffix || '';
    return $('[data-list' + suffix + '="' + name + '"]');
}

function listNextName(list, attr) {
    var max = -1,
        prefix = attr.substr(5);

    if (typeof list == 'string')
        list = listMatch(list);

    listGetItems(list).each(function () {
        var name = this.getAttribute(attr),
            value;

        if (!name.startsWith(prefix))
            return;

        value = parseInt(name.replace(new RegExp('^' + prefix), ''), 10);

        if (!isNaN(value))
            max = Math.max(max, value);
    });

    return prefix + (max + 1);
}

function listRegisterItemCallback(name, callback) {
    LIST_CALLBACKS[name] = callback;
}

function listSay(list, text, type) {
    var $listmesg;

    if (typeof list == 'string')
        list = listMatch(list);

    $listmesg = list.find('[data-listmesg="' + list.attr('data-list') + '"]')
        .removeClass('success info warning error')
        .text(text || '')
        .toggle(text ? true : false);

    if (type)
        $listmesg.addClass(type);
}

function listSetupFilterInit() {
    var $filters;

    $filters = $('[data-listfilter]').each(function () {
        this.setAttribute('autocomplete', 'off');
        this._lastValue = '';

        // Get associated list
        this._list = $body.find('[data-list="' + this.getAttribute('data-listfilter') + '"]').get(0);
    });

    if ($filters.length > 0) {
        $body.on('keyup', '[data-listfilter]', function (e) {
            var listId = e.target.getAttribute('data-listfilter');

            if (e.which == 27)
                e.target.value = '';

            if (!e._force && e.target.value == e.target._lastValue)
                return;

            if (LIST_TIMEOUTS[listId])
                clearTimeout(LIST_TIMEOUTS[listId]);

            // Update list content
            LIST_TIMEOUTS[listId] = setTimeout(function () {
                listUpdate($(e.target._list), e.target.value);
                e.target._lastValue = e.target.value;
            }, 200);
        });
    }
}

function listSetupInit() {
    return $.Deferred(function ($deferred) {
        var $deferreds = [],
            $lists;

        $lists = $('[data-list]').each(function () {
            $deferreds.push(listInit(this));
        });

        $.when.apply(null, $deferreds).then(function () { $deferred.resolve(); });

        if ($lists.length > 0) {
            $body.on('click', '[data-listmesg] a', function (e) {
                var $target = $(e.target),
                    event = {
                        type: 'keyup',
                        _force: true
                    };

                if ($target.attr('href') == "#reset")
                    event.which = 27;
                else if ($target.attr('href') != "#retry")
                    return;

                $('[data-listfilter="' + $target.closest('[data-list]').attr('data-list') + '"]')
                    .trigger(event);

                e.preventDefault();
                e.stopImmediatePropagation();
            });
        }
    }).promise();
}

function listSetupMoreInit() {
    var $more;

    $more = $('[data-listmore]');

    if ($more.length > 0) {
        $body.on('click', '[data-listmore]', function (e) {
            var listId = e.target.getAttribute('data-listmore');
            listUpdate(listId, listMatch(listId, 'filter').val(), listGetCount(listId));
        });
    }
}

function listUpdate(list, listFilter, offset) {
    var query,
        timeout,
        url;

    offset = parseInt(offset, 10) || 0;

    if (typeof list == 'string')
        list = listMatch(list);

    // Set query timeout
    timeout = setTimeout(function () {
        overlayCreate('loader', {
            message: $.t('main.mesg_loading')
        });
    }, 500);

    // Empty list if not appending entries
    if (offset === 0) {
        listEmpty(list);
        listMatch(list.attr('data-list'), 'more').attr('disabled', 'disabled');
    }

    // Request data
    url = list.opts('list').url;

    query = {
        url: urlPrefix + '/api/v1/' + url,
        type: 'GET',
        data: {
            offset: offset,
            limit: LIST_FETCH_LIMIT
        }
    };

    if (listFilter)
        query.data.filter = 'glob:*' + listFilter + '*';

    return $.ajax(query).done(function (data, status, xhr) { /*jshint unused: true */
        var $item,
            name = list.attr('data-list'),
            namespace,
            records = parseInt(xhr.getResponseHeader('X-Total-Records'), 10),
            i;

        if (!data || data instanceof Array && data.length === 0) {
            namespace = list.opts('list').messages || 'item';

            if (listFilter) {
                listSay(list, $.t(namespace + '.mesg_load_nomatch'), 'warning');

                $(document.createElement('a')).appendTo(list.find('[data-listmesg]'))
                    .attr('href', '#reset')
                    .text($.t('list.labl_reset'));
            } else {
                listSay(list, $.t(namespace + '.mesg_none'), 'info');
            }

            return;
        }

        listSay(list, null);

        for (i in data) {
            if (typeof data[i] == 'string') {
                listAppend(list)
                    .attr('data-itemname', data[i])
                    .find('.name').text(data[i]);

                continue;
            }

            $item = listAppend(list)
                .attr('data-itemid', data[i].id);

            $item.find('.name').text(data[i].name);
            $item.find('.desc').text(data[i].description || $.t('main.mesg_no_description'));
            $item.find('.date span').text(moment(data[i].modified).format(TIME_DISPLAY));

            if (!data[i].description)
                $item.find('.desc').addClass('placeholder');

            // Execute item callback if any
            if (LIST_CALLBACKS[name])
                LIST_CALLBACKS[name]($item, data[i]);
        }

        listUpdateCount(list, records);

        if (listGetCount(list) < records)
            listMatch(list.attr('data-list'), 'more').removeAttr('disabled');
        else
            listMatch(list.attr('data-list'), 'more').attr('disabled', 'disabled');
    }).fail(function () {
        listEmpty(list);
        listSay(list, $.t('list.mesg_load_error'), 'error');

        $(document.createElement('a')).appendTo(list.find('[data-listmesg]'))
            .attr('href', '#retry')
            .text($.t('list.labl_retry'));
    }).always(function () {
        if (timeout)
            clearTimeout(timeout);

        overlayDestroy('loader');
    });
}

function listUpdateCount(list, count) {
    if (typeof list == 'string')
        list = listMatch(list);

    // Update list count
    list.find('h1 .count').text((count ? count : listGetCount(list)) || '');
}

// Register setup callbacks
setupRegister(SETUP_CALLBACK_TERM, listSetupInit);
setupRegister(SETUP_CALLBACK_TERM, listSetupFilterInit);
setupRegister(SETUP_CALLBACK_TERM, listSetupMoreInit);
