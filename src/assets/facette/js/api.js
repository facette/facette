function apiTransformRequest(data) {
    delete data.type;
    return JSON.stringify(data);
}

function apiInterceptList(data) {
    data.resource.$totalRecords = parseInt(data.headers('X-Total-Records'), 10);
    return data.resource;
}

app.factory('bulk', function($resource) {
    return $resource('api/v1/bulk', null, {
        exec: {
            method: 'POST',
            isArray: true,
            transformRequest: apiTransformRequest
        }
    });
});

app.factory('catalog', function($resource) {
    return $resource('api/v1/catalog/:type/:name', {
        type: '@type',
        name: '@name'
    }, {
        list: {
            method: 'GET',
            params: {
                type: '@type',
                name: null
            },
            isArray: true,
            interceptor: {
                response: apiInterceptList
            }
        }
    });
});

app.factory('expand', function($resource) {
    return $resource('api/v1/expand', null, {
        fetch: {
            method: 'POST',
            isArray: true,
            transformRequest: apiTransformRequest
        },
    });
});

app.factory('info', function($resource) {
    return $resource('api/v1/', null, {
        get: {
            method: 'GET'
        }
    });
});

app.factory('library', function($resource) {
    return $resource('api/v1/library/:type/:id', {
        type: '@type',
        id: '@id'
    }, {
        append: {
            method: 'POST',
            params: {
                type: '@type',
                id: null
            },
            transformRequest: apiTransformRequest
        },
        collectionTree: {
            method: 'GET',
            params: {
                type: 'collections',
                id: 'tree'
            },
            isArray: true
        },
        count: {
            method: 'HEAD',
            params: {
                type: '@type',
                id: null
            },
            interceptor: {
                response: apiInterceptList
            }
        },
        delete: {
            method: 'DELETE',
            params: {
                type: '@type',
                id: '@id'
            }
        },
        get: {
            method: 'GET',
            params: {
                type: '@type',
                id: '@id'
            }
        },
        list: {
            method: 'GET',
            params: {
                type: '@type',
                id: null
            },
            isArray: true,
            interceptor: {
                response: apiInterceptList
            }
        },
        update: {
            method: 'PUT',
            params: {
                type: '@type',
                id: '@id'
            },
            transformRequest: apiTransformRequest
        },
        search: {
            method: 'POST',
            params: {
                type: 'search',
                id: null,
                limit: '@limit'
            },
            isArray: true,
            transformRequest: apiTransformRequest
        }
    });
});

app.factory('libraryAction', function($resource) {
    return $resource('api/v1/library/:action', {
        action: '@action'
    }, {
        parse: {
            method: 'POST',
            params: {
                action: 'parse'
            },
            isArray: true
        },
        search: {
            method: 'POST',
            params: {
                action: 'search'
            },
            isArray: true
        }
    });
});

app.factory('plots', function($resource) {
    return $resource('api/v1/plots', null, {
        fetch: {
            method: 'POST',
            transformRequest: apiTransformRequest
        }
    });
});

app.factory('providers', function($resource) {
    return $resource('api/v1/providers/:id', {
        id: '@id'
    }, {
        append: {
            method: 'POST',
            params: {
                id: null
            }
        },
        delete: {
            method: 'DELETE',
            params: {
                id: '@id'
            }
        },
        get: {
            method: 'GET',
            params: {
                id: '@id'
            }
        },
        list: {
            method: 'GET',
            params: {
                id: null
            },
            isArray: true
        },
        update: {
            method: 'PATCH',
            params: {
                id: '@id'
            }
        }
    });
});

app.factory('providersAction', function($resource) {
    return $resource('api/v1/providers/:id/:action', {
        action: '@action'
    }, {
        refresh: {
            method: 'POST',
            params: {
                id: '@id',
                action: 'refresh'
            }
        }
    });
});
