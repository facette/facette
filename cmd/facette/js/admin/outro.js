
function adminSetupInit() {
    // Hide pane steps
    $('[data-step]').hide();
}

if (String(window.location.pathname).startsWith(urlPrefix + '/admin/')) {
    // Register setup callbacks
    setupRegister(SETUP_CALLBACK_INIT, adminSetupInit);
    setupRegister(SETUP_CALLBACK_TERM, adminCollectionSetupTerminate);
    setupRegister(SETUP_CALLBACK_TERM, adminGraphSetupTerminate);
    setupRegister(SETUP_CALLBACK_TERM, adminGroupSetupTerminate);
    setupRegister(SETUP_CALLBACK_TERM, adminScaleSetupTerminate);
    setupRegister(SETUP_CALLBACK_TERM, adminUnitSetupTerminate);
    setupRegister(SETUP_CALLBACK_TERM, adminCatalogSetupTerminate);
}
