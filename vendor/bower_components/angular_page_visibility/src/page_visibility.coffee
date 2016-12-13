angular.module('angular-page-visibility', [])
  .factory('$pageVisibility', [ '$rootScope', '$document', ($rootScope, $document)->
    pageVisibility = $rootScope.$new()

    getVisibilityKeys = ->
      if typeof($document.prop('hidden')) != 'undefined'
        [ 'hidden', 'visibilitychange' ]
      else if typeof($document.prop('mozHidden')) != 'undefined'
        [ 'mozHidden', 'mozvisibilitychange' ]
      else if typeof($document.prop('msHidden')) != 'undefined'
        [ 'msHidden', 'msvisibilitychange' ]
      else if typeof($document.prop('webkitHidden')) != 'undefined'
        [ 'webkitHidden', 'webkitvisibilitychange' ]

    return pageVisibility unless getVisibilityKeys()

    [hiddenKey, visibilityChagedKey] = getVisibilityKeys()

    $document.on(visibilityChagedKey, ->
      if $document.prop(hiddenKey)
        pageVisibility.$broadcast('pageBlurred')
      else
        pageVisibility.$broadcast('pageFocused')
    )

    pageVisibility
  ])
