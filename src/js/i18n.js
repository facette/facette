
/* i18n */

function i18nSetupInit() {
    // Load messages resource and initialize i18n support
    return $.ajax({
        url: urlPrefix + '/static/messages.json',
        type: 'GET',
    }).pipe(function (data) {
        $.i18n.init({
            lng: 'en',
            resStore: {
                en: {
                    translation: data
                }
            }
        });
    });
}

// Register setup callbacks
setupRegister(SETUP_CALLBACK_INIT, i18nSetupInit);
