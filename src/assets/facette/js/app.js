var app = angular.module('facette', [
    'angucomplete-alt',
    'angular-inview',
    'angular-page-visibility',
    'as.sortable',
    'bw.paging',
    'facette.ui.color',
    'facette.ui.column',
    'facette.ui.dialog',
    'facette.ui.form',
    'facette.ui.graph',
    'facette.ui.include',
    'facette.ui.list',
    'facette.ui.menu',
    'facette.ui.message',
    'facette.ui.notify',
    'facette.ui.pane',
    'facette.ui.search',
    'facette.ui.tab',
    'facette.ui.tabindex',
    'ngDialog',
    'ngResource',
    'ngRoute',
    'ngSanitize',
    'pascalprecht.translate',
    'ui.bootstrap.datetimepicker',
    'ui.dateTimeInput',
    'ui.select',
    'ui.tree'
]);

app.config(function($httpProvider, $locationProvider, $resourceProvider, $routeProvider, $translateProvider,
    treeConfig) {

    // Set HTML5-styled URLs
    $locationProvider
        .html5Mode(true)
        .hashPrefix('!');

    // Register application routes
    $routeProvider
        .when('/browse/', {
            templateUrl: 'templates/browse/search.html',
            controller: 'BrowseSearchController'
        })
        .when('/browse/:section/:id', {
            templateUrl: 'templates/browse/graphs.html',
            controller: 'BrowseGraphController',
            reloadOnSearch: false
        })
        .when('/show/collections/:id/:index', {
            templateUrl: 'templates/show/collections.html',
            controller: 'ShowCollectionController'
        })
        .when('/show/graphs/:id', {
            templateUrl: 'templates/show/graphs.html',
            controller: 'ShowGraphController'
        })
        .when('/admin/info/', {
            templateUrl: 'templates/admin/info.html',
            controller: 'AdminInfoController'
        })
        .when('/admin/collections/', {
            templateUrl: 'templates/admin/list-library.html',
            controller: 'AdminListController',
            _type: 'collections'
        })
        .when('/admin/collections/:id', {
            templateUrl: 'templates/admin/edit-collections.html',
            controller: 'AdminEditCollectionController'
        })
        .when('/admin/graphs/', {
            templateUrl: 'templates/admin/list-library.html',
            controller: 'AdminListController',
            _type: 'graphs'
        })
        .when('/admin/graphs/:id', {
            templateUrl: 'templates/admin/edit-graphs.html',
            controller: 'AdminEditGraphController'
        })
        .when('/admin/sourcegroups/', {
            templateUrl: 'templates/admin/list-library.html',
            controller: 'AdminListController',
            _type: 'sourcegroups'
        })
        .when('/admin/sourcegroups/:id', {
            templateUrl: 'templates/admin/edit-groups.html',
            controller: 'AdminEditGroupController',
            _type: 'sourcegroups'
        })
        .when('/admin/metricgroups/', {
            templateUrl: 'templates/admin/list-library.html',
            controller: 'AdminListController',
            _type: 'metricgroups'
        })
        .when('/admin/metricgroups/:id', {
            templateUrl: 'templates/admin/edit-groups.html',
            controller: 'AdminEditGroupController',
            _type: 'metricgroups'
        })
        .when('/admin/providers/', {
            templateUrl: 'templates/admin/list-providers.html',
            controller: 'AdminListController',
            _type: 'providers'
        })
        .when('/admin/providers/:id', {
            templateUrl: 'templates/admin/edit-providers.html',
            controller: 'AdminEditProviderController'
        })
        .when('/admin/origins/', {
            templateUrl: 'templates/admin/list-catalog.html',
            controller: 'AdminListController',
            _type: 'origins'
        })
        .when('/admin/sources/', {
            templateUrl: 'templates/admin/list-catalog.html',
            controller: 'AdminListController',
            _type: 'sources'
        })
        .when('/admin/metrics/', {
            templateUrl: 'templates/admin/list-catalog.html',
            controller: 'AdminListController',
            _type: 'metrics'
        })
        .when('/admin/', {
            redirectTo: '/admin/collections/'
        })
        .when('/', {
            redirectTo: '/browse/'
        })
        .otherwise({
            templateUrl: 'templates/error/404.html',
            controller: 'ErrorController'
        });

    // Add global API calls error handler
    $httpProvider.interceptors.push(function($q, $rootScope) {
        return {
            request: function(config) {
                return config;
            },
            requestError: function(response) {
                return $q.reject(response);
            },
            response: function(response) {
                return response;
            },
            responseError: function(response) {
                if (response.status >= 400 && response.status != 404) {
                    $rootScope.$emit('Notify', response.data && response.data.message ?
                        response.data.message : 'an unhandled error has occurred', {type: 'error'});

                    if (response.status == 403) {
                        // Force read-only recheck
                        $rootScope.checkReadOnly();
                    }
                }

                return $q.reject(response);
            }
        };
    });

    // Don't strip trailing slash on requests
    $resourceProvider.defaults.stripTrailingSlashes = false;

    // Set up translation
    $translateProvider
        .useMessageFormatInterpolation()
        .useSanitizeValueStrategy(null)
        .useStaticFilesLoader({
            prefix: '/assets/js/locales/',
            suffix: '.json'
        })
        .registerAvailableLanguageKeys(['en', 'fr'], {
            'en_*': 'en',
            'fr_*': 'fr',
            '*': 'en'
        })
        .determinePreferredLanguage();

    // Set up tree defaults
    treeConfig.defaultCollapsed = true;
});

app.run(function($anchorScroll, $browser, $location, $pageVisibility, $rootScope, $route, $timeout, $translate, $window,
    info, ngDialog, storage) {

    $rootScope.baseURL = $browser.baseHref();
    $rootScope.title = null;
    $rootScope.hasFocus = true;

    $rootScope.stateOK = stateOK;
    $rootScope.stateLoading = stateLoading;
    $rootScope.stateError = stateError;

    // Handle page title
    $rootScope.setTitle = function(parts) {
        if (!parts) {
            $rootScope.title = null;
            return;
        }

        $translate(parts).then(function(data) {
            var title = [];

            for (var i = 0, n = parts.length; i < n; i++) {
                title.push(data[parts[i]]);
            }

            $rootScope.title = title.join(' – ');
        });
    };

    // Define unique repeat tracking
    // (see https://github.com/a5hik/ng-sortable/issues/128)
    $rootScope.unique = function(idx, id) {
        return idx + id;
    };

    // Handle page unload events
    function unloadFunc() {
        return 'mesg.page_unload';
    }

    $rootScope.preventedUnload = false;

    $rootScope.preventUnload = function(state) {
        if (state === $rootScope.preventedUnload) {
            return;
        }

        $rootScope.preventedUnload = state;
        angular.element($window)[state ? 'on' : 'off']('beforeunload', unloadFunc);
    };

    // Handle modal display
    $rootScope.showModal = function(data, callback) {
        ngDialog.open({
            template: 'templates/common/dialog.html',
            controller: function($scope) {
                $scope.data = data;

                $timeout(function() {
                    var element = angular.element('.ngdialog-content input:first');
                    if (element.length > 0) {
                        element.select();
                    } else {
                        angular.element('.ngdialog-content .validate').focus();
                    }
                }, 50);
            }
        }).closePromise.then(function(scope) {
            scope.$dialog.remove();

            if ([undefined, '$document', '$escape'].indexOf(scope.value) != -1) {
                return;
            }

            callback(scope.value);
        });
    };

    $rootScope.handleDatetimeKey = function(e) {
        var element = angular.element(e.target);
        if (element.attr('tabindex') !== undefined) {
            element.trigger('click');
        }
    };

    // Handle alt modifier
    $rootScope.handleAlt = function(e) {
        if (e.which != 18) {
            return;
        }

        $rootScope.altMode = e.type == 'keydown';
    };

    // Handle sidebar toggle
    $rootScope.sidebarCollapse = storage.get('global-sidebar', 'collapsed', false);

    $rootScope.toggleSidebar = function() {
        $rootScope.sidebarCollapse = !$rootScope.sidebarCollapse;
        storage.set('global-sidebar', 'collapsed', $rootScope.sidebarCollapse);
    };

    // Handle local preferences reset
    $rootScope.resetLocalPrefs = function() {
        localStorage.clear();
        location.reload();
    };

    // Handle read-only instance
    $rootScope.readOnly = false;

    $rootScope.checkReadOnly = function() {
        info.get(null, function(data) {
            if (data.read_only) {
                $rootScope.readOnly = true;
            }
        });
    };

    $rootScope.checkReadOnly();

    // Extend location
    $location.skipReload = function() {
        var currentRoute = $route.current;

        var unbind = $rootScope.$on('$locationChangeSuccess', function(e) {
            $route.current = currentRoute;
            unbind();
        });

        return $location;
    };

    // Attach events
    $pageVisibility.$on('pageFocused', function(e) {
        $rootScope.hasFocus = true;
    });

    $pageVisibility.$on('pageBlurred', function(e) {
        $rootScope.hasFocus = false;
    });

    $rootScope.$on('$includeContentLoaded', function(e, src) {
        if (!src.endsWith('/layout.html')) {
            return;
        }

        // Scroll to anchor
        if (!$anchorScroll.yOffset) {
            $anchorScroll.yOffset = angular.element('article header').height();
        }

        $anchorScroll();
    });

    $rootScope.$on("$locationChangeStart", function(e) {
        if (!$rootScope.preventedUnload) {
            return;
        }

        var path = $location.path(),
            search = $location.search();

        e.preventDefault();

        $rootScope.showModal({
            type: dialogTypeConfirm,
            message: unloadFunc(),
            labels: {
                validate: 'label.leave_page'
            },
            danger: true
        }, function(data) {
            if (data === undefined) {
                return;
            }

            $rootScope.preventUnload(false);
            $location.path(path).search(search);
        });
    });

    $rootScope.$on("$routeChangeSuccess", function() {
        $rootScope.altMode = false;
        $rootScope.inAdmin = $location.path().startsWith('/admin');
        $rootScope.preventUnload(false);
    });

    $rootScope.$on('$translateLoadingEnd', function() {
        $rootScope.loaded = true;
    });

    angular.element($window).on('resize', function() {
        if ($rootScope.resizeTimeout) {
            $timeout.cancel($rootScope.resizeTimeout);
            $rootScope.resizeTimeout = null;
        }

        $rootScope.resizeTimeout = $timeout(function() {
            if (angular.element($window).width() <= sidebarCollapseWith) {
                if (!$rootScope.sidebarCollapse) {
                    $rootScope.sidebarCollapse = true;
                    $rootScope.sidebarCollapseAuto = true;
                }
            } else if ($rootScope.sidebarCollapseAuto && $rootScope.sidebarCollapse) {
                $rootScope.sidebarCollapse = false;
                $rootScope.sidebarCollapseAuto = false;
            }
        }, 500);
    });

    angular.element($window).trigger('resize');
});
