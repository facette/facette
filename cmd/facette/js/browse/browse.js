
function browsePrint() {
    // Force graphs load then trigger print
    graphHandleQueue(true).then(function () {
        window.print();
    });
}

function browseSetRange(e) {
    var $target = $(e.target),
        $overlay,
        href = $target.attr('href');

    // Prevent event from being triggered from a graph item
    if ($target.closest('[data-graph]').length > 0)
        return;

    if (href == '#set-global-range') {
        $target.next('.menu').toggle();
    } else if (href == '#range-custom') {
        $overlay = overlayCreate('time', {
            callbacks: {
                validate: function () {
                    $('[data-graph]').each(function () {
                        var $item = $(this);

                        $.extend($item.data('options'), {
                            time: moment($overlay.find('input[name=time]').val()).format(TIME_RFC3339),
                            range: $overlay.find('input[name=range]').val()
                        });

                        graphDraw($item, !$item.inViewport());
                    });
                }
            }
        });

        $overlay.find('input[name=time]').appendDtpicker({
            closeOnSelected: true,
            current: null,
            firstDayOfWeek: 1,
            minuteInterval: 10,
            todayButton: false
        });

        $('a[href=#set-global-range] + .menu').hide();

        e.stopImmediatePropagation();
    } else if (href && href.indexOf('#range-') === 0) {
        $('[data-graph]').each(function () {
            var $item = $(this);

            $.extend($item.data('options'), {range: '-' + href.substr(7)});
            delete $item.data('options').time;

            graphDraw($item, !$item.inViewport());
        });

        $target.closest('.menu').hide();
    } else if ($target.closest('.menu').length === 0) {
        $('a[href=#set-global-range] + .menu').hide();
        return;
    } else {
        return;
    }

    e.preventDefault();
    e.stopPropagation();
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

function browseToggleLegend(e) {
    $(e.target).toggleClass('icon-toggle-off icon-toggle-on active');
    $('[data-graph] a[href=#toggle-legend]').trigger('click');
}
