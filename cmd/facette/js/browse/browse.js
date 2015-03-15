
function browsePrint() {
    // Force graphs load then trigger print
    graphHandleQueue(true).then(function () {
        window.print();
    });
}

function browseSetRefresh(e) {
    overlayCreate('prompt', {
        message: $.t('main.labl_refresh_interval'),
        callbacks: {
            validate: function (data) {
                if (!data)
                    return;

                data = parseInt(data, 10);
                if (isNaN(data)) {
                    consoleToggle($.t('main.mesg_invalid_refresh_interval'));
                    return;
                }

                $('[data-graph]').each(function () {
                    var $item = $(this);
                    $.extend($item.data('options'), {refresh_interval: data});
                    graphDraw($item, !$item.inViewport());
                });

                // Set refresh interval UI display
                $(e.target)
                    .addClass('value')
                    .html('<span>' + data + '</span>');
            }
        }
    });
}
