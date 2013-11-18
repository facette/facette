
function adminSetupInit() {
    // Hide pane steps
    $('[data-step]').hide();
}

if (String(window.location.pathname).startsWith('/admin/')) {
    // Register links
    linkRegister('reload', adminReloadServer);

    // Register setup callbacks
    setupRegister(SETUP_CALLBACK_INIT, adminSetupInit);
    setupRegister(SETUP_CALLBACK_TERM, adminGraphSetupTerminate);
    setupRegister(SETUP_CALLBACK_TERM, adminCollectionSetupTerminate);
    setupRegister(SETUP_CALLBACK_TERM, adminGroupSetupTerminate);
}
