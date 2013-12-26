
function adminSetupInit() {
    // Hide pane steps
    $('[data-step]').hide();
}

if (String(window.location.pathname).startsWith('/admin/')) {
    // Register links
    linkRegister('reload', function () {
        overlayCreate('confirm', {
            message: $.t('main.mesg_reload'),
            callbacks: {
                validate: adminReloadServer
            },
            labels: {
                validate: {
                    text: $.t('main.labl_reload'),
                    style: 'danger'
                }
            }
        });
    });

    // Register setup callbacks
    setupRegister(SETUP_CALLBACK_INIT, adminSetupInit);
    setupRegister(SETUP_CALLBACK_TERM, adminGraphSetupTerminate);
    setupRegister(SETUP_CALLBACK_TERM, adminCollectionSetupTerminate);
    setupRegister(SETUP_CALLBACK_TERM, adminGroupSetupTerminate);
    setupRegister(SETUP_CALLBACK_TERM, adminCatalogSetupTerminate);
}
