
if (String(window.location.pathname).startsWith(urlPrefix + '/browse/')) {
    // Register links
    linkRegister('print', browsePrint);

    // Register setup callbacks
    setupRegister(SETUP_CALLBACK_TERM, browseCollectionSetupTerminate);
}
