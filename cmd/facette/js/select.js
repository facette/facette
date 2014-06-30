
/* Select */

var $selectTemplate;

function selectHandleChange(e) {
    var $target,
        $select;

    if (e._select)
        return;

    $target = $(e.target);
    $select = $target.closest('[data-select]');

    // Set current item
    $select.find('[data-menuitem="' + $target.val() + '"]').trigger({
        type: 'click',
        _init: e._init || false
    });
}

function selectHandleClick(e) {
    var $target = $(e.target),
        $item = $target.closest('[data-menuitem]'),
        $select = $target.closest('[data-select]');

    if ($item.length === 0) {
        menuToggle($select.attr('data-select'));
    } else {
        $select.find('[data-selectlabel]')
            .text($item.text());

        $select.children('select')
            .val($item.data('value'))
            .trigger({
                type: 'change',
                _init: e._init || false,
                _select: true
            });

        menuToggle($select.attr('data-select'), false);
    }
}

function selectHandleFocus(e) {
    // Toggle menu display
    if (e.type == 'focusin') {
        $body.on('keydown keyup', selectHandleKey);
    } else {
        $body.off('keydown keyup', selectHandleKey);
    }
}

function selectHandleKey(e) {
    var $select;

    if (e.type == 'keydown') {
        if ([EVENT_KEY_UP, EVENT_KEY_LEFT, EVENT_KEY_DOWN, EVENT_KEY_RIGHT].indexOf(e.which) != -1)
            e.preventDefault();

        return;
    }

    $select = $('[data-selectlabel]:focus').closest('[data-select]');

    if (menuMatch($select.attr('data-select')).is(':visible'))
        return;

    if (e.which == EVENT_KEY_UP || e.which == EVENT_KEY_LEFT)
        $select.find('[data-menuitem="' + $select.children('select').val() + '"]')
            .prev('[data-menuitem]').trigger('click');
    else if (e.which == EVENT_KEY_DOWN || e.which == EVENT_KEY_RIGHT)
        $select.find('[data-menuitem="' + $select.children('select').val() + '"]')
            .next('[data-menuitem]').trigger('click');
}

function selectInit(element) {
    var $element = $(element),
        $menu,
        $select;

    $select = $selectTemplate.clone().insertBefore(element)
        .attr('data-select', $element.attr('data-select'));

    // Create new menu
    $menu = menuCreate(element.getAttribute('data-select')).appendTo($select)
        .attr('data-menu', $select.attr('data-select'));

    $element.detach().appendTo($select)
        .removeAttr('data-select')
        .hide();

    // Set label focusable
    $select.find('[data-selectlabel]')
        .attr('tabindex', 0);

    selectUpdate(element);
}

function selectUpdate(element) {
    var $element = $(element),
        $select = $element.closest('[data-select]'),
        $menu = $select.find('[data-menu]');

    menuEmpty($menu);

    $element.children('option').each(function () {
        menuAppend($menu)
            .attr('data-menuitem', this.value)
            .data('value', this.value)
            .text($(this).text());
    });

    $element.trigger({
        type: 'change',
        _init: true
    });

    // Make width consistent
    if ($menu.width() > $select.width())
        $select.width($menu.width());
    else
        $menu.css('min-width', $select.width());
}

function selectSetupInit() {
    // Get main objects
    $selectTemplate = $('[data-select=template]').detach();

    // Initialize select items
    $('[data-select]').each(function () { selectInit(this); });
}

// Attach events
$body
    .on('change', '[data-select] select', selectHandleChange)
    .on('click', '[data-select] .selectlabel, [data-select] [data-menuitem]', selectHandleClick)
    .on('focusin focusout', '[data-select]', selectHandleFocus);

// Register setup callbacks
setupRegister(SETUP_CALLBACK_INIT, selectSetupInit);
