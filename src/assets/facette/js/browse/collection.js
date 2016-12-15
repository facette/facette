app.factory('BrowseCollection', function($timeout, library, storage) {
    return {
        injectTree: function(scope) {
            library.list({
                type: 'collections',
                fields: 'id,name,parent,options.title',
                kind: 'raw',
                expand: 1
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

            // Register watchers
            scope.$watch('collections', function(newValue, oldValue) {
                if (angular.equals(newValue, oldValue)) {
                    return;
                }

                // Rerieve existing tree state
                var state = storage.get('browse-collection_tree', 'state', {});

                $timeout(function() {
                    var trees = angular.element('#collections-tree .tree'),
                        baseMargin = parseInt(trees.first().find('.treelabel').css('padding-left'), 10);

                    trees.each(function(index, item) {
                        var tree = angular.element(item);

                        tree.children('.treeitem').children('.treelabel').css({
                            paddingLeft: parseInt(tree.closest('.treeitem').children('.treelabel')
                                .css('padding-left'), 10) + baseMargin
                        });
                    });

                    // Restore tree state
                    if (state) {
                        angular.element('#collections-tree .treelabel').each(function(index, item) {
                            var label = angular.element(item),
                                href = label.attr('href');

                            if (state[href.substr(href.lastIndexOf('/') + 1)] === true) {
                                label.children('.toggle').trigger('click');
                            }
                        });
                    }

                    scope.collectionsLoaded = true;
                }, 250);
            });
        },
        saveTreeState: function() {
            var state = {};
            angular.element('#collections-tree [collapsed]').each(function(index, item) {
                var href = angular.element(item).children('.treelabel').attr('href'),
                    id = href.substr(href.lastIndexOf('/') + 1);

                state[id] = item.getAttribute('collapsed') == 'false';
            });

            storage.set('browse-collection_tree', 'state', state);
        }
    };
});
