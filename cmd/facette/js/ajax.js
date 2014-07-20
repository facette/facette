/* Ajax */

$.ajaxSetup({
    complete: function (xhr) {
        var data = {};

        if (xhr.status < 400)
            return;

        try {
            data = JSON.parse(xhr.responseText);
        } catch (e) {}

        consoleToggle(data.message || $.t('main.mesg_unknown_error'));
    }
});
