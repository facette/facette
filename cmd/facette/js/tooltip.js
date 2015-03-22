
/* Tooltip */

var TOOLTIP_ACTIVE    = null,
    TOOLTIP_CALLBACKS = {},

    $tooltipTemplate;

function tooltipCreate(name, toggleCallback) {
    if (toggleCallback)
        TOOLTIP_CALLBACKS[name] = toggleCallback;

    // Remove any prexisting tooltip with the same name
    $('[data-tooltip="' + name + '"]').remove();

    return tooltipToggle($tooltipTemplate.clone().attr('data-tooltip', name), true);
}

function tooltipHandleClick(e) {
    if ($(e.target).closest('[data-tooltip]').length === 0)
        tooltipToggle(null, false);
}

function tooltipHandleKey(e) {
    if (e.which == 27)
        tooltipToggle(null, false);
}

function tooltipMatch(name) {
    return $('[data-tooltip="' + name + '"]');
}

function tooltipSetupInit() {
    // Get main objects
    $tooltipTemplate = $('[data-tooltip=template]').detach();
}

function tooltipToggle(tooltip, state) {
    // Apply on all tooltips if none specified
    if (!tooltip) {
        $('[data-tooltip]').each(function () { tooltipToggle($(this), state); });
        return;
    }

    if (typeof tooltip == 'string')
        tooltip = tooltipMatch(tooltip);

    state = typeof state == 'boolean' ? state : tooltip.is(':hidden');

    if (state) {
        if (!TOOLTIP_ACTIVE) {
            // Attach tooltip events
            $body
                .on('keydown', tooltipHandleKey)
                .on('click', tooltipHandleClick);
        }

        TOOLTIP_ACTIVE = tooltip;
    } else if (TOOLTIP_ACTIVE && tooltip.get(0) == TOOLTIP_ACTIVE.get(0)) {
        TOOLTIP_ACTIVE = null;

        // Detach tooltip events
        $body
            .off('keydown', tooltipHandleKey)
            .off('click', tooltipHandleClick);
    }

    if (TOOLTIP_CALLBACKS[tooltip.attr('data-tooltip')])
        TOOLTIP_CALLBACKS[tooltip.attr('data-tooltip')](state);

    return tooltip.toggle(state);
}

// Attach events
$window
    .on('resize', function () { tooltipToggle(null, false); });

// Register setup callbacks
setupRegister(SETUP_CALLBACK_INIT, tooltipSetupInit);
