app.factory('storage', function() {
    var data = {};

    var service = {
        get: function(namespace, key, fallback) {
            if (!data[namespace]) {
                service.load(namespace);
            }

            return data[namespace][key] ? data[namespace][key] : fallback;
        },

        set: function(namespace, key, value) {
            if (!data[namespace]) {
                service.load(namespace);
            }

            data[namespace][key] = value;
            service.save(namespace);
        },

        load: function(namespace) {
            data[namespace] = JSON.parse(localStorage.getItem(namespace)) || {};
        },

        save: function(namespace) {
            localStorage.setItem(namespace, JSON.stringify(data[namespace]));
        }
    };

    return service;
});
