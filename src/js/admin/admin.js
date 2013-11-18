
function adminReloadServer() {
    ADMIN_RELOAD_TIMEOUT = setTimeout(function () {
        overlayCreate('loader', {
            message: $.t('main.mesg_server_loading')
        });
    }, 500);

    return $.ajax({
        url: '/reload',
        type: 'GET'
    }).then(function () {
        if (ADMIN_RELOAD_TIMEOUT)
            clearTimeout(ADMIN_RELOAD_TIMEOUT);

        overlayDestroy('loader');

        window.location.reload(true);
    });
}
