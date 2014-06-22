
function adminCatalogSetupTerminate() {
    // Register admin panes
    paneRegister('catalog-list', function () {
        // Register links
        linkRegister('show-info', function (e) {
            var $target = $(e.target),
                $item = $target.closest('[data-itemname]'),
                type = $target.closest('[data-list]').attr('data-list'),
                name = $item.attr('data-itemname');

            $.ajax({
                url: urlPrefix + '/api/v1/catalog/' + type + '/' + name,
                type: 'GET'
            }).pipe(function (data) {
                var $tooltip = tooltipCreate('info', function (state) {
                    $item.toggleClass('action', state);
                }).appendTo($body)
                    .css({
                        top: $target.offset().top,
                        left: $target.offset().left
                    });

                switch (type) {
                    case 'origins':
                        $tooltip.html(
                            '<span class="label">' + $.t('main.labl_connector') + '</span> ' +
                                data.connector
                        );

                        break;

                    case 'sources':
                        $tooltip.html(
                            '<span class="label">' + $.t('main.labl_origins') + '</span> ' +
                                data.origins.join(', ')
                        );

                        break;

                    case 'metrics':
                        $tooltip.html(
                            '<span class="label">' + $.t('main.labl_origins') + '</span> ' +
                                data.origins.join(', ') + '<br>' +
                            '<span class="label">' + $.t('main.labl_sources') + '</span> ' +
                                data.sources.join(', ')
                        );

                        break;
                }
            });
        });
    });
}
