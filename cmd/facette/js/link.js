
/* Link */

var LINK_CALLBACKS = {};

function linkRegister(fragments, callback) {
    $.each($.map(fragments.split(' '), $.trim), function (i, fragment) { /*jshint unused: true */
        LINK_CALLBACKS[fragment] = callback;
    });
}

function linkHandleClick(e) {
    var fragment;

    for (fragment in LINK_CALLBACKS) {
        if (!e.target.href || !e.target.href.endsWith('#' + fragment))
            continue;

        if (e.target.getAttribute('disabled') != 'disabled')
            LINK_CALLBACKS[fragment](e);

        e.preventDefault();
    }
}

// Attach events
$body
    .on('click', 'a', linkHandleClick);
