
/* Menu */

var MENU_ACTIVE         = null,
    MENU_SCROLL_LOCKED  = false,
    MENU_SCROLL_TIMEOUT = null,

    $menuTemplate,
    $menuTemplateItem;

function menuAppend(menu) {
    if (typeof menu == 'string')
        menu = menuMatch(menu);

    return $menuTemplateItem.clone().appendTo(menu.find('[data-menucntr]'));
}

function menuCreate(name) {
    return $menuTemplate.clone()
        .attr('data-menu', name)
        .hide();
}

function menuEmpty(menu) {
    if (typeof menu == 'string')
        menu = menuMatch(menu);

    menu.find('[data-menuitem]').remove();
}

function menuHandleClick(e) {
    if ($(e.target).closest('[data-menu], [data-input], [data-select]').length === 0)
        menuToggle(null, false);
}

function menuHandleKey(e) {
    var $item,
        $items,
        $menucntr,
        $selected,
        position;

    // Stop if non-handled key
    if ([EVENT_KEY_ENTER, EVENT_KEY_TAB, EVENT_KEY_DOWN, EVENT_KEY_UP].indexOf(e.which) == -1 || !MENU_ACTIVE) {
        // Hide menu if <Escape> pressed
        if (e.which == EVENT_KEY_ESCAPE)
            menuToggle(MENU_ACTIVE, false);

        return;
    }

    e.preventDefault();

    if (MENU_SCROLL_TIMEOUT)
        clearTimeout(MENU_SCROLL_TIMEOUT);

    // Get current menu and selected item
    $items    = MENU_ACTIVE.find('[data-menuitem]');
    $selected = $items.filter('.selected');

    if (e.which == EVENT_KEY_ENTER || e.which == EVENT_KEY_TAB) {
        if (e.which == EVENT_KEY_TAB && (e.shiftKey || MENU_ACTIVE.closest('[data-input]').length === 0))
            return;

        // Execute action
        $selected.trigger({
            type: 'click',
            which: 1
        });

        menuToggle(MENU_ACTIVE, false);
        return;
    } else if (e.which == EVENT_KEY_UP) {
        MENU_SCROLL_LOCKED = true;

        // Select previous item
        $item = $selected.prevAll('[data-menuitem]:not(.noselect)').first();

        if ($item.length === 0)
            $item = $items.not('.noselect').last();
    } else if (e.which == EVENT_KEY_DOWN) {
        MENU_SCROLL_LOCKED = true;

        // Select next item
        $item = $selected.nextAll('[data-menuitem]:not(.noselect)').first();

        if ($item.length === 0)
            $item = $items.not('.noselect').first();
    }

    $item.addClass('selected').siblings().removeClass('selected');

    // Update scroll position
    position = $item.position();

    if (position) {
        $menucntr = MENU_ACTIVE.find('[data-menucntr]');

        if (position.top + $item.outerHeight(true) > $menucntr.innerHeight()) {
            $menucntr.scrollTop($menucntr.scrollTop() + position.top - $menucntr.innerHeight() +
                $item.outerHeight(true) + ($menucntr.outerHeight() - $menucntr.innerHeight()));
        } else if (position.top < 0) {
            $menucntr.scrollTop($menucntr.scrollTop() + position.top);
        }
    }

    if (MENU_SCROLL_LOCKED) {
        MENU_SCROLL_TIMEOUT = setTimeout(function () {
            MENU_SCROLL_LOCKED = false;
            $body.on('mousemove', '[data-menuitem]', menuHandleMouse);
        }, 200);
    }
}

function menuHandleMouse(e) {
    var $item;

    if (MENU_SCROLL_LOCKED)
        return;

    // Update selection state
    if (e.type == 'mouseenter' || e.type == 'mousemove') {
        $item = $(e.target).closest('[data-menuitem]');

        if (!$item.hasClass('noselect'))
            $item.addClass('selected').siblings().removeClass('selected');

        if (e.type == 'mousemove')
            $body.off('mousemove', '[data-menuitem]', menuHandleMouse);
    } else {
        $(e.target).closest('[data-menu]').find('[data-menuitem].selected')
            .removeClass('selected');
    }
}

function menuMatch(name) {
    return $('[data-menu="' + name + '"]');
}

function menuSay(menu, text, type) {
    if (typeof menu == 'string')
        menu = menuMatch(menu);

    menu.find('[data-menumesg]')
        .attr('data-menumesg', type || 'info')
        .text(text || '')
        .toggle(text ? true : false);

    if (text)
        menuToggle(menu, true);
}

function menuSetupInit() {
    // Get main objects
    $menuTemplate     = $('[data-menu=template]');
    $menuTemplateItem = $menuTemplate.find('[data-menuitem=template]').detach();

    $menuTemplate.detach();
}

function menuToggle(menu, state) {
    // Apply on all menus if none specified
    if (!menu) {
        $('[data-menu]').each(function () { menuToggle($(this), state); });
        return;
    }

    if (typeof menu == 'string')
        menu = menuMatch(menu);

    state = typeof state == 'boolean' ? state : menu.is(':hidden');

    if (state) {
        // Hide current visible menu and set the new one
        if (MENU_ACTIVE && menu.get(0) != MENU_ACTIVE.get(0)) {
            menuToggle(MENU_ACTIVE, false);
        } else if (!MENU_ACTIVE) {
            // Attach menu events
            $body
                .on('keydown', menuHandleKey)
                .on('mouseenter', '[data-menuitem]', menuHandleMouse)
                .on('mouseleave', '[data-menu]', menuHandleMouse);
        }

        MENU_ACTIVE = menu;

        // Unselect previously selected items
        menu.find('[data-menuitem].selected').removeClass('selected');
    } else if (MENU_ACTIVE && menu.get(0) == MENU_ACTIVE.get(0)) {
        MENU_ACTIVE = null;

        // Detach menu events
        $body
            .off('keydown', menuHandleKey)
            .off('mouseenter', '[data-menuitem]', menuHandleMouse)
            .off('mouseleave', '[data-menu]', menuHandleMouse);
    }

    menu.toggle(state);
}

// Attach events
$window
    .on('resize', function () { menuToggle(null, false); });

$body
    .on('click', menuHandleClick);

// Register setup callbacks
setupRegister(SETUP_CALLBACK_INIT, menuSetupInit);
