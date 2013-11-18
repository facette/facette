
/* Overlay */

var OVERLAY_TEMPLATES = {},

    $overlay;

function overlayCreate(type, args) {
    var $input,
        $item;

    if (!OVERLAY_TEMPLATES[type]) {
        console.error("Unable find `" + type + "' overlay");
        return;
    }

    $item = OVERLAY_TEMPLATES[type].clone().appendTo($overlay.show())
        .data('args', args)
        .on('click', 'button', function (e) {
            var $item,
                args;

            if (e.target.name != 'cancel' && e.target.name != 'validate')
                return;

            $item = $(e.target).closest('[data-overlay]');
            args  = $item.data('args');

            if (args && args.callbacks) {
                if (args.callbacks[e.target.name])
                    args.callbacks[e.target.name](type == 'prompt' ? $item.find('input[type=text]').val() : null);

                if (args.callbacks.terminate)
                    args.callbacks.terminate();
            }

            overlayDestroy($item);
        });

    $body.on('keydown', overlayHandleKey);

    if (args) {
        if (args.message)
            $item.find('.message').text(args.message);

        if (args.labels) {
            $.each(args.labels, function (name, info) {
                var $label = $item.find('button[name=' + name + ']');

                if (info.text)
                    $label.text(info.text);

                if (info.style)
                    $label.addClass(info.style);
            });
        }

        if (type == 'prompt') {
            $input = $item.find('input[type=text]:first');

            if (args.value)
                $input.val(args.value);

            setTimeout(function () { $input.select(); }, 0);
        }
    }

    return $item;
}

function overlayDestroy(overlay) {
    if (typeof overlay == 'string')
        overlay = overlayMatch(overlay);

    if (overlay.length === 0)
        return;

    overlay.remove();

    if ($overlay.find('[data-overlay]').length === 0)
        $overlay.hide();

    $body.off('keydown', overlayHandleKey);
}

function overlayHandleKey(e) {
    if (e.which != 13 && e.which != 27)
        return;

    $overlay.children('[data-overlay]').each(function () {
        $(this).find('button[name=' + (e.which == 13 || e.which == 27 &&
            this.getAttribute('data-overlay') == 'alert' ? 'validate' : 'cancel') + ']').trigger('click');
    });

    e.preventDefault();
}

function overlayMatch(name) {
    return $overlay.children('[data-overlay=' + name + ']');
}

function overlaySetupInit() {
    // Initialize overlay
    $overlay = $('#overlay').hide();

    $overlay.find('.box, .loader').each(function () {
        var $item = $(this);
        OVERLAY_TEMPLATES[$item.attr('data-overlay')] = $item.detach();
    });

    if ($('[data-wait]').length > 0)
        adminReloadServer();
}

// Register setup callbacks
setupRegister(SETUP_CALLBACK_INIT, overlaySetupInit);
