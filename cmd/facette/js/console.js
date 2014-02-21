
/* Console */

var $console;

function consoleSetupInit() {
    // Get main objects
    $console = $('#console');

    // Register links
    linkRegister('console-close', function () {
        consoleToggle(null);
    });

    consoleToggle(null);
}

function consoleToggle(message) {
    if (message && $console.is(':hidden'))
        $console.show();

    $console.children('.message').text(message);

    $console.animate({top: message ? 0 : $console.outerHeight(true) * -1}, 200, function () {
        if (!message)
            $console.hide();
    });
}

// Register setup callbacks
setupRegister(SETUP_CALLBACK_INIT, consoleSetupInit);
