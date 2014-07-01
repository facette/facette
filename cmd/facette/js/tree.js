
/* Tree */

function treeAppend(tree) {
    // Append new item
    return tree.data('template').clone().appendTo(tree.children('.treecntr'));
}

function treeEmpty(tree) {
    tree.children('.treecntr').empty();
}

function treeInit(element) {
    return $.Deferred(function ($deferred) {
        var $item = $(element),
            $template;

        if (!$.contains(document.documentElement, element)) {
            $deferred.resolve();
            return;
        }

        $template = $item.find('.treetmpl:first').removeClass('treetmpl');
        $item.data('template', $template);
        $template.detach();

        $item.children('.placeholder').hide();

        // Initialize tree content
        treeUpdate($item).then(function () { $deferred.resolve(); });
    }).promise();
}

function treeMatch(name) {
    return $('[data-tree=' + name + ']');
}

function treeSetupInit() {
    return $.Deferred(function ($deferred) {
        var $deferreds = [];

        $('[data-tree]').each(function () {
            $deferreds.push(treeInit(this));
        });

        $.when.apply(null, $deferreds).then(function () { $deferred.resolve(); });
    }).promise();
}

function treeUpdate(tree) {
    var opts,
        query = {},
        timeout;

    if (typeof tree == 'string')
        tree = treeMatch(tree);

    // Set query timeout
    timeout = setTimeout(function () {
            overlayCreate('loader', {
            message: $.t('main.mesg_loading')
        });
    }, 500);

    // Request data
    opts = tree.opts('tree');

    if (opts.base)
        query.parent = opts.base;

    return $.ajax({
        url: urlPrefix + '/api/v1/' + opts.url,
        type: 'GET',
        data: query
    }).pipe(function (data) {
        var $item,
            i;

        treeEmpty(tree);
        tree.children('.placeholder').toggle(data.length === 0);

        if (data.length === 0)
            return;

        for (i in data) {
            $item = treeAppend(tree)
                .attr('data-treeitem', data[i].id)
                .attr('title', data[i].name);

            if (data[i].has_children)
                $item.addClass('folded');

            $item.find('.name').text(data[i].name);
        }

        for (i in data) {
            if (data[i].parent === null)
                continue;

            $('[data-treeitem=' + data[i].id + ']').detach()
                .appendTo($('[data-treeitem=' + data[i].parent + ']').children('.treecntr').hide());
        }

        treeUpdatePadding(tree);
    }).fail(function () {
        treeEmpty(tree);
    }).always(function () {
        if (timeout)
            clearTimeout(timeout);

        overlayDestroy('loader');
    });
}

function treeUpdatePadding(tree) {
    var $containers,
        marginBase,
        i = 0;

    if (typeof tree == 'string')
        tree = treeMatch(tree);

    $containers = tree.find('.treecntr');
    marginBase  = Math.abs(parseInt($containers.first().find('.treelabl').css('margin-left'), 10));

    $containers.each(function () {
        var margin = -(marginBase * i);

        $(this).closest('.treeitem').find('.treelabl').css({
            marginLeft: margin,
            marginRight: margin,
            paddingLeft: Math.abs(margin)
        });

        i += 1;
    });
}

// Register setup callbacks
setupRegister(SETUP_CALLBACK_INIT, treeSetupInit);
