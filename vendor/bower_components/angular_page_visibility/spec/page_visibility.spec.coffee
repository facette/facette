describe 'Page Visibility', ->
  $document = null
  beforeEach module 'angular-page-visibility'
  beforeEach module ($provide)->
    $document =
      attrs: {}
      events: {}
      _setAttr: (key, val)->
        @attrs[key] = val
      attr: (key)->
        @attrs[key]
      prop: (key)->
        @attrs[key]
      on: (eventName, cb)->
        @events[eventName] = cb
      _trigger: (eventName)->
        @events[eventName] && @events[eventName]()

    $provide.value('$document', $document)
    null

  ensurePageVisibilityEventsWith = (hiddenKey, visibilityChagedKey)->
    describe "when document visibility is controlled by `#{hiddenKey}`", ->
      it '$broadcast-s `pageFocused` when page turn visible', ->
        inject (_$document_)->
          $document._setAttr(hiddenKey, false)

        inject ($pageVisibility)->
          onFocused = sinon.spy()
          $pageVisibility.$on('pageFocused', onFocused)

          $document._trigger(visibilityChagedKey)

          expect(onFocused.called).toBe(true)

      it '$broadcast-s `pageBlurred` when page turns invisible', ->
        inject (_$document_)->
          $document._setAttr(hiddenKey, true)

        inject ($pageVisibility)->
          onBlurred = sinon.spy()
          $pageVisibility.$on('pageBlurred', onBlurred)

          $document._trigger(visibilityChagedKey)

          expect(onBlurred.called).toBe(true)

  ensurePageVisibilityEventsWith('hidden', 'visibilitychange')
  ensurePageVisibilityEventsWith('mozHidden', 'mozvisibilitychange')
  ensurePageVisibilityEventsWith('msHidden', 'msvisibilitychange')
  ensurePageVisibilityEventsWith('webkitHidden', 'webkitvisibilitychange')

  it 'does nothing if the browser does not support page visibility API', ->
    inject ($pageVisibility)->

