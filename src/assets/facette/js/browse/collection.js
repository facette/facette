app.factory('BrowseCollection', function(library) {
    return {
        injectTree: function(scope) {
            library.list({
                type: 'collections',
                fields: 'id,name,parent,options.title',
                kind: 'raw'
            }, function(data) {
                var tree = {},
                    collections = [];

                angular.forEach(data, function(item) {
                    // Set name to title if any
                    if (item.options && item.options.title) {
                        item.name = item.options.title;
                    }

                    tree[item.id] = angular.extend(tree[item.id] ? tree[item.id] : {children: []}, item);

                    // Set collection link
                    tree[item.id].href = 'browse/collections/' + item.id;

                    if (!item.parent) {
                        return;
                    }

                    if (tree[item.parent]) {
                        tree[item.parent].children.push(tree[item.id]);
                    } else {
                        tree[item.parent] = {children: [tree[item.id]]};
                    }
                });

                angular.forEach(tree, function(item) {
                    if (!scope.id && !item.parent || scope.id && item.parent == scope.id) {
                        collections.push(item);
                    }
                });

                scope.collections = collections;
            });
        }
    };
});
