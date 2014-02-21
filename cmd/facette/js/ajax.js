/* Ajax */

$.ajaxSetup({
    complete: function (xhr) {
        var data,
            message;

        if (xhr.status < 400)
            return;

        try {
            data = JSON.parse(xhr.responseText);
            message = data.message;
        } catch (e) {}

        consoleToggle(message || $.t('main.mesg_unknown_error'));
    }
});
