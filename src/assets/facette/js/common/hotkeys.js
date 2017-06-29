app.factory('globalHotkeys', function($location, $translate, hotkeys) {
    var combos = {
        'g h':  {description: 'label.goto_home', path: '/'},
        'g a':  {description: 'label.goto_admin', path: '/admin/'},
        'g c':  {description: 'label.goto_list_collections', path: '/admin/collections/'},
        'g g':  {description: 'label.goto_list_graphs', path: '/admin/graphs/'},
        'g s':  {description: 'label.goto_list_sourcegroups', path: '/admin/sourcegroups/'},
        'g m':  {description: 'label.goto_list_metricgroups', path: '/admin/metricgroups/'},
        'g o':  {description: 'label.goto_list_origins', path: '/admin/origins/'},
        'g S':  {description: 'label.goto_list_sources', path: '/admin/sources/'},
        'g M':  {description: 'label.goto_list_metrics', path: '/admin/metrics/'},
        'g p':  {description: 'label.goto_list_providers', path: '/admin/providers/'},
        'g i':  {description: 'label.goto_info', path: '/admin/info/'},
        'c c':  {description: 'label.goto_create_collections', path: '/admin/collections/add'},
        'c C':  {description: 'label.goto_create_collections_link', path: '/admin/collections/link'},
        'c g':  {description: 'label.goto_create_graphs', path: '/admin/graphs/add'},
        'c G':  {description: 'label.goto_create_graphs_link', path: '/admin/graphs/link'},
        'c s':  {description: 'label.goto_create_sourcegroups', path: '/admin/sourcegroups/add'},
        'c m':  {description: 'label.goto_create_metricgroups', path: '/admin/metricgroups/add'},
        'c p':  {description: 'label.goto_create_providers', path: '/admin/providers/add'}
    }

    function handleGotoHotkey(e, hotkey) {
        if (combos[hotkey.combo[0]]) {
            $location.path(combos[hotkey.combo[0]].path);
        }
    }

    return {
        register: function(scope) {
            var hk = hotkeys.bindTo(scope);
            angular.forEach(combos, function(attrs, combo) {
                hk.add({combo: combo, description: attrs.description, callback: handleGotoHotkey});
            });
        }
    };
});
