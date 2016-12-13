(function() {
  angular.module('angular-page-visibility', []).factory('$pageVisibility', [
    '$rootScope', '$document', function($rootScope, $document) {
      var getVisibilityKeys, hiddenKey, pageVisibility, visibilityChagedKey, _ref;
      pageVisibility = $rootScope.$new();
      getVisibilityKeys = function() {
        if (typeof ($document.prop('hidden')) !== 'undefined') {
          return ['hidden', 'visibilitychange'];
        } else if (typeof ($document.prop('mozHidden')) !== 'undefined') {
          return ['mozHidden', 'mozvisibilitychange'];
        } else if (typeof ($document.prop('msHidden')) !== 'undefined') {
          return ['msHidden', 'msvisibilitychange'];
        } else if (typeof ($document.prop('webkitHidden')) !== 'undefined') {
          return ['webkitHidden', 'webkitvisibilitychange'];
        }
      };
      if (!getVisibilityKeys()) {
        return pageVisibility;
      }
      _ref = getVisibilityKeys(), hiddenKey = _ref[0], visibilityChagedKey = _ref[1];
      $document.on(visibilityChagedKey, function() {
        if ($document.prop(hiddenKey)) {
          return pageVisibility.$broadcast('pageBlurred');
        } else {
          return pageVisibility.$broadcast('pageFocused');
        }
      });
      return pageVisibility;
    }
  ]);

}).call(this);

//# sourceMappingURL=page_visibility.js.map
