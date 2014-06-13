
function adminScaleGetData() {
    var $pane = paneMatch('scale-edit');

    return {
        name: $pane.find('input[name=scale-name]').val(),
        description: $pane.find('textarea[name=scale-desc]').val(),
        value: parseFloat($pane.find('input[name=scale-value]').val())
    };
}

function adminScaleSetupTerminate() {
    // Register admin panes
    paneRegister('scale-list', function () {
        adminItemHandlePaneList('scale');
    });

    paneRegister('scale-edit', function () {
        var scaleId = paneMatch('scale-edit').opts('pane').id || null;

        // Attach events
        $body
            .on('click', 'button', function (e) {
                var $target = $(e.target);

                switch (e.target.name) {
                case 'step-cancel':
                    window.location = urlPrefix + '/admin/scales/';
                    break;

                case 'step-save':
                    adminItemHandlePaneSave($target.closest('[data-pane]'), scaleId, 'scale', adminScaleGetData);
                    break;
                }
            });

        // Load scale data
        if (scaleId === null)
            return;

        itemLoad(scaleId, 'scales').pipe(function (data) {
            var $pane = paneMatch('scale-edit');

            $pane.find('input[name=scale-name]').val(data.name);
            $pane.find('textarea[name=scale-desc]').val(data.description);
            $pane.find('input[name=scale-value]').val(data.value);
        });
    });
}
