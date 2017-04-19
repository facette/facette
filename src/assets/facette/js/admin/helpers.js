app.factory('adminHelpers', function($rootScope, catalog, library, providers) {
    var catalogSections = [
            'origins',
            'sources',
            'metrics'
        ],

        librarySections = [
            'collections',
            'graphs',
            'sourcegroups',
            'metricgroups'
        ];

    return {
        getFactory: function(scope) {
            if (catalogSections.indexOf(scope.section) != -1) {
                return catalog;
            } else if (librarySections.indexOf(scope.section) != -1) {
                return library;
            } else if (scope.section == 'providers') {
                return providers;
            }

            return null;
        }
    };
});
