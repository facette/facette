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
    'facette.ui.menu',
    'facette.ui.message',
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
        .when('/admin/:section/', {
            templateUrl: function(params) {
                var category;
                if (catalogSections.indexOf(params.section) != -1) {
                    category = 'catalog';
                } else if (librarySections.indexOf(params.section) != -1) {
                    category = 'library';
                } else {
                    category = params.section;
                }

                return 'templates/admin/list-' + category + '.html';
            },
            controller: 'AdminListController'
        })
        .when('/admin/collections/:id', {
            templateUrl: 'templates/admin/edit-collections.html',
            controller: 'AdminEditCollectionController'
        })
        .when('/admin/graphs/:id', {
            templateUrl: 'templates/admin/edit-graphs.html',
            controller: 'AdminEditGraphController'
        })
        .when('/admin/sourcegroups/:id', {
            templateUrl: 'templates/admin/edit-groups.html',
            controller: 'AdminEditGroupController',
            _type: 'sourcegroups'
        })
        .when('/admin/metricgroups/:id', {
            templateUrl: 'templates/admin/edit-groups.html',
            controller: 'AdminEditGroupController',
            _type: 'metricgroups'
        })
        .when('/admin/scales/:id', {
            templateUrl: 'templates/admin/edit-scales.html',
            controller: 'AdminEditScaleController'
        })
        .when('/admin/units/:id', {
            templateUrl: 'templates/admin/edit-units.html',
            controller: 'AdminEditUnitController'
        })
        .when('/admin/providers/:id', {
            templateUrl: 'templates/admin/edit-providers.html',
            controller: 'AdminEditProviderController'
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
                if (response.status >= 400) {
                    $rootScope.setError(response.data && response.data.message ?
                        response.data.message : 'an unhandled error has occurred');
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
        .registerAvailableLanguageKeys(['en'], {
            'en_*': 'en',
            '*': 'en'
        })
        .determinePreferredLanguage();

    // Set up tree defaults
    treeConfig.defaultCollapsed = true;
});

app.run(function($anchorScroll, $browser, $location, $pageVisibility, $rootScope, $timeout, $translate, $window,
    ngDialog) {

    $rootScope.baseURL = $browser.baseHref();
    $rootScope.title = null;
    $rootScope.hasFocus = true;

    $rootScope.stateOK = stateOK;
    $rootScope.stateLoading = stateLoading;
    $rootScope.stateError = stateError;

    // Handle page title
    $rootScope.setTitle = function(parts) {
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

    // Handle sidebar toggle
    $rootScope.toggleSidebar = function() {
        $rootScope.sidebarOpen = !$rootScope.sidebarOpen;
    };

    // Handle error message
    $rootScope.setError = function(content) {
        $rootScope.error = content;
        $rootScope.errorActive = true;
        $timeout(function() { $rootScope.resetError(); }, 5000);
    };

    $rootScope.resetError = function() {
        $rootScope.errorActive = false;
        $timeout(function() { $rootScope.error = null; }, 250);
    };

    $rootScope.resetError();

    // Handle local preferences reset
    $rootScope.resetLocalPrefs = function() {
        localStorage.clear();
        location.reload();
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
        var path = $location.path(),
            search = $location.search();

        if ($rootScope.preventedUnload) {
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
        }
    });

    $rootScope.$on("$routeChangeSuccess", function() {
        $rootScope.preventUnload(false);
    });

    $rootScope.$on('$translateLoadingEnd', function() {
        $rootScope.loaded = true;
    });
});
