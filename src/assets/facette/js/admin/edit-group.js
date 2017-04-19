app.controller('AdminEditGroupController', function($q, $route, $routeParams, $scope, $translate, adminEdit, catalog) {
    $scope.section = $route.current.$$route._type;
    $scope.id = $routeParams.id;

    // Define scope functions
    $scope.cancel = function(force) { adminEdit.cancel($scope, force); };
    $scope.delete = function() { adminEdit.delete($scope, {id: $scope.id, name: $scope.itemRef.name}); };
    $scope.reset = function() { adminEdit.reset($scope); };
    $scope.save = function() { adminEdit.save($scope); };
    $scope.remove = function(list, entry) { adminEdit.remove($scope, list, entry); };

    $scope.selectPattern = function(data) {
        if (!data || !data.originalObject || !data.originalObject.name) {
            return;
        }

        angular.extend($scope.pattern, {
            type: $scope.patternTypes[0],
            value: data.originalObject.name
        });
    };

    $scope.setPattern = function() {
        var pattern;

        switch ($scope.pattern.type.value) {
        case groupPatternGlob:
            pattern = patternPrefixGlob + $scope.pattern.value;
            break;

        case groupPatternRegexp:
            pattern = patternPrefixRegexp + $scope.pattern.value;
            break;

        default:
            pattern = $scope.pattern.value;
        }

        if (!$scope.item.patterns) {
            $scope.item.patterns = [];
        }

        if ($scope.pattern.index !== undefined) {
            $scope.item.patterns[$scope.pattern.index] = pattern;
        } else {
            $scope.item.patterns.push(pattern);
        }

        $scope.resetPattern();
    };

    $scope.editPattern = function(entry) {
        var idx = $scope.item.patterns.indexOf(entry);
        if (idx == -1) {
            return;
        }

        var focus = '#value';

        if (entry.indexOf(patternPrefixGlob) === 0) {
            $scope.pattern = {type: $scope.patternTypes[1], value: entry.substr(patternPrefixGlob.length)};
        } else if (entry.indexOf(patternPrefixRegexp) === 0) {
            $scope.pattern = {type: $scope.patternTypes[2], value: entry.substr(patternPrefixRegexp.length)};
        } else {
            $scope.pattern = {type: $scope.patternTypes[0], value: entry};
            $scope.$broadcast('angucomplete-alt:changeInput', 'value', entry);
            focus += '_value';
        }

        $scope.pattern.index = idx;

        $scope.$applyAsync(function() { angular.element(focus).select(); });
    };

    $scope.testPattern = function(pattern) {
        var limit = 10,
            defer = $q.defer();

        $q.all([
            $translate(['label.patterns_matches', 'label.patterns_matches_total', 'label.patterns_matches_none']),
            catalog.list({
                type: $scope.section == 'sourcegroups' ? 'sources' : 'metrics',
                limit: limit,
                filter: pattern
            }).$promise
        ]).then(function(data) {
            if (data[1].$totalRecords === 0) {
                data[1].push(data[0]['label.patterns_matches_none']);
            }

            var content = '<span>' + data[0]['label.patterns_matches'] + '</span><br>\n' + data[1].join('<br>\n');

            if (data[1].$totalRecords > limit) {
                content += '<br>\n…<br>\n<span>' + data[0]['label.patterns_matches_total'] + '</span> ' +
                    data[1].$totalRecords;
            }

            defer.resolve(content);
        });

        return defer.promise;
    };

    $scope.resetPattern = function() {
        $scope.pattern = {
            type: $scope.patternTypes[0],
            value: null
        };

        $scope.$broadcast('angucomplete-alt:clearInput', 'value');
        $scope.$applyAsync(function() { angular.element('#value_value').focus(); });
    };

    // Register watchers
    adminEdit.watch($scope);

    // Initialize scope
    adminEdit.load($scope, function() {
        var type = $scope.section.replace(/groups$/, 's');

        $scope.patternTypes = [
            {name: 'Single', value: groupPatternSingle},
            {name: 'Glob', value: groupPatternGlob},
            {name: 'Regexp', value: groupPatternRegexp}
        ];

        $scope.patternValues = function(term) {
            var defer = $q.defer();

            catalog.list({type: type, filter: 'glob:*' + term + '*'}).$promise.then(function(data) {
                var result = [];
                angular.forEach(data, function(name) {
                    result.push({name: name});
                });
                defer.resolve(result);
            });

            return defer.promise;
        };

        $scope.resetPattern();
    });
});
