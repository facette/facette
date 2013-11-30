
/* Scroll */

var SCROLL_TIMEOUT = null;

function scrollSetupTerminate() {
    // Attach events
    $('.scrollarea').on('scroll', function (e) {
        var $area = $(this);

        if (SCROLL_TIMEOUT)
            clearTimeout(SCROLL_TIMEOUT);

        $area.addClass('scroll');

        SCROLL_TIMEOUT = setTimeout(function () {
            $area.removeClass('scroll');
        }, 500);
    });
}

// Register setup callbacks
setupRegister(SETUP_CALLBACK_TERM, scrollSetupTerminate);
