
if (String(window.location.pathname).startsWith('/browse/')) {
    // Register links
    linkRegister('print', browsePrint);

    // Register setup callbacks
    setupRegister(SETUP_CALLBACK_TERM, browseCollectionSetupTerminate);
}
