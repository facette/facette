
function adminUnitGetData() {
    var $pane = paneMatch('unit-edit');

    return {
        name: $pane.find('input[name=unit-name]').val(),
        description: $pane.find('textarea[name=unit-desc]').val(),
        label: $pane.find('input[name=unit-label]').val(),
        type: parseInt($pane.find('input[name=unit-type]:checked').val(), 10)
    };
}

function adminUnitSetupTerminate() {
    // Register admin panes
    paneRegister('unit-list', function () {
        adminItemHandlePaneList('unit');
    });

    paneRegister('unit-edit', function () {
        var unitId = paneMatch('unit-edit').opts('pane').id || null;

        // Attach events
        $body
            .on('click', 'button', function (e) {
                var $target = $(e.target);

                switch (e.target.name) {
                case 'step-cancel':
                    window.location = urlPrefix + '/admin/units/';
                    break;

                case 'step-save':
                    adminItemHandlePaneSave($target.closest('[data-pane]'), unitId, 'unit', adminUnitGetData);
                    break;
                }
            });

        // Load unit data
        if (unitId === null)
            return;

        itemLoad(unitId, 'units').pipe(function (data) {
            var $pane = paneMatch('unit-edit');

            $pane.find('input[name=unit-name]').val(data.name);
            $pane.find('textarea[name=unit-desc]').val(data.description);
            $pane.find('input[name=unit-label]').val(data.label);
            $pane.find('input[name=unit-type][value=' + data.type + ']').prop('checked', true);
        });
    });
}
