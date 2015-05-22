
if (locationPath.startsWith(urlPrefix + '/browse/')) {
    // Register links
    linkRegister('print', browsePrint);

    // Register setup callbacks
    setupRegister(SETUP_CALLBACK_TERM, browseCollectionSetupTerminateTree);

    if (locationPath.startsWith(urlPrefix + '/browse/collections/'))
        setupRegister(SETUP_CALLBACK_TERM, browseCollectionSetupTerminate);

    if (locationPath.startsWith(urlPrefix + '/browse/graphs/'))
        setupRegister(SETUP_CALLBACK_TERM, browseGraphSetupTerminate);
}
