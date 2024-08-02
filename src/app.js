//@ts-check

import { lineNumbers, drawSelection, highlightActiveLine } from '@codemirror/view'
import { keymap, dropCursor, highlightActiveLineGutter, highlightSpecialChars } from '@codemirror/view'
import { syntaxHighlighting, indentOnInput, defaultHighlightStyle, bracketMatching, foldGutter } from '@codemirror/language'
import { autocompletion, closeBrackets } from '@codemirror/autocomplete'
import { highlightSelectionMatches } from '@codemirror/search'
import { defaultKeymap, indentWithTab } from '@codemirror/commands'
import { noctisLilac } from 'thememirror'

import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { javascript } from '@codemirror/lang-javascript'
import { python } from '@codemirror/lang-python'
import { go } from '@codemirror/lang-go'
import { java } from '@codemirror/lang-java'
import { cpp } from '@codemirror/lang-cpp'
import { php } from '@codemirror/lang-php'
import { html } from '@codemirror/lang-html'
import { sql } from '@codemirror/lang-sql'

import * as Y from 'yjs'
import { yCollab, yUndoManagerKeymap } from 'y-codemirror.next'
import { WebsocketProvider } from 'y-websocket'

/**
 * @param {string} codeContent
 * @param {(() => import("@codemirror/state").Extension) | null} languageExtension
 * @param {RemoteCollaboration?} collab
 */
function createEditorState(codeContent, languageExtension, collab = null) {
  if (!languageExtension) {
    languageExtension = () => []
  }

  if (!codeContent) {
    codeContent = ""
  }

  let startState = EditorState.create({
    doc: codeContent,
    extensions: [
      // dracula,
      noctisLilac,
      // EditorState.lineSeparator.of("\n"),
      lineNumbers(),
      drawSelection(),
      foldGutter(),
      // rectangularSelection(),
      highlightActiveLine(),
      highlightActiveLineGutter(),
      dropCursor(),
      // crosshairCursor(),
      highlightSelectionMatches(),
      highlightSpecialChars(),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      indentOnInput(),
      bracketMatching(),
      closeBrackets(),
      autocompletion(),
      keymap.of([
        ...defaultKeymap,
        indentWithTab,
        ...yUndoManagerKeymap,
      ]),
      languageExtension(),
      collab ? yCollab(collab.sharedContent, collab.provider.awareness) : [],
    ],
  })

  return startState
}

/**
 * @param {EditorView} view
 * @param {EditorState} freshState
 */
function editEditorView(view, freshState) {
  view.setState(freshState)
}

/**
 * @param {Element} editorCanvas
 * @param {EditorState} startState
 * @return {EditorView} 
 */
function createEditorView(editorCanvas, startState) {
  editorCanvas.innerHTML = ''

  let view = new EditorView({
    state: startState,
    parent: editorCanvas,
  })

  view.dom.style.minHeight = "300px"
  view.dom.style.outline = "0"

  return view
}

/**
 * @param {HTMLSelectElement} el
 */
function selectLanguageFromDOM(el) {
  let languageExtension = getLanguageExtensionFromLabel(el.value)

  console.log(`[Info] [reconfigureCodeEditor] languageExtension = `, languageExtension)

  return languageExtension
}

/**
 * @param {string} languageLabel
 */
function getLanguageExtensionFromLabel(languageLabel) {
  let languageExtension

  switch (languageLabel) {
    case "js":
      languageExtension = javascript
      break
    case "python":
      languageExtension = python
      break;
    case "go":
      languageExtension = go
      break;
    case "java":
      languageExtension = java
      break;
    case "cpp":
      languageExtension = cpp
      break;
    case "php":
      languageExtension = php
      break;
    case "html":
      languageExtension = html
      break;
    case "sql":
      languageExtension = sql
      break;
    default:
      languageExtension = javascript
      break;
  }

  return languageExtension
}

/**
 * @param {EditorView} view
 * @param {any} languageExtension
 * @param {RemoteCollaboration?} collab
 */
function reconfigureCodeEditor(view, languageExtension, collab = null) {
  let freshState = createEditorState(view.state.doc.toString(), languageExtension, collab)
  editEditorView(view, freshState)
}

/**
 * @param {number} postID
 */
function getPostElement(postID) {
  return document.getElementById(`user-post__${postID}`)
}

/**
 * @param {HTMLElement?} postHtmlElement
 * @return {HTMLSelectElement?}
 */
function getEditorLanguageElement(postHtmlElement) {
  if (!postHtmlElement)
    return null

  return postHtmlElement.querySelector(".code-editor__header select")
}

/**
 * @param {HTMLElement?} postHtmlElement
 */
function getEditorBodyElement(postHtmlElement) {
  if (!postHtmlElement)
    return null

  return postHtmlElement.getElementsByClassName("code-editor__body")[0]
}

function getCreationPostID() {
  let defaultPostID = 0
  let post = APP.emptyPost

  if (!post) {
    console.log("[Warning] No Data for empty 'Post' found")
    console.log("[Info] Post Creation may not work as expected !")

    return defaultPostID
  }

  return post.PostID
}


/**
 * @typedef {Object} Post
 * @property {number} PostID
 * @property {string} CodeSnipet
 * @property {string} LanguageCode
 */

/**
 * @typedef {Map<number, EditorView>} ViewsMap
 */

/**
 * From a postID, the function scan the DOM to find an element matching the corresponding DOM id
 * @param {Post} post
 * @returns {EditorView?}
 */
function buildPostDOM(post) {
  let postID = post.PostID

  let postHtmlElement = getPostElement(postID)
  if (!postHtmlElement) {
    console.log("[Error] postHtmlElement == null, abording Code Editor binding ...")
    return null
  }

  let codeEditorElement = getEditorBodyElement(postHtmlElement)
  let codeEditorLanguageElement = getEditorLanguageElement(postHtmlElement)

  if (!codeEditorElement || !codeEditorLanguageElement) {
    console.log("[Error] codeEditor(Language)Element == null, abording Code Editor binding ...")
    return null
  }

  let posterCode = post.CodeSnipet

  let startState = createEditorState(posterCode, null)
  let view = createEditorView(codeEditorElement, startState)

  let languageLabel = post.LanguageCode
  if (!languageLabel) {
    languageLabel = "js"

    console.log("[Warning] Language for code editor was empty ! Value set to default : ", languageLabel)
  }

  let languageExtension = getLanguageExtensionFromLabel(languageLabel)
  reconfigureCodeEditor(view, languageExtension)


  /** @param {any} el */
  storage.standardEditorEventListenerFunction = function(el) {
    let languageExtension = selectLanguageFromDOM(el.target)

    reconfigureCodeEditor(view, languageExtension)
    console.log("[Info] Successfully switched to Standard language: ", el.target.value)
  }

  // @ts-ignore
  codeEditorLanguageElement.value = languageLabel
  codeEditorLanguageElement.addEventListener("change", storage.standardEditorEventListenerFunction)

  console.log("[Info] Created a post")

  return view
}

/**
 * @param {Post} post
 * @returns {EditorView?}
 */
function generateOriginalPostDOM(post) {
  if (!post || post.PostID <= 0) {
    console.log("[Warning] No Original Post Found !")
    return null
  }

  let view = buildPostDOM(post)

  if (view)
    console.log("[Info] Created Original Post (look above)")

  return view
}

/**
 * @param {Post[]} posts
 * @returns {ViewsMap}
 */
function generateAnswerPostsDOM(posts) {
  /** @type {ViewsMap} */
  let views = new Map()

  if (!posts) {
    console.log("[Warning] No Answer Posts Found !")
    return views
  }

  for (let i = 0; i < posts.length; i++) {
    if (posts[i].PostID <= 0) {
      console.log("[Warning] An answer Post have an expected ID")
      continue
    }

    let view = buildPostDOM(posts[i])
    if (view === null)
      continue

    views.set(posts[i].PostID, view)
  }

  if (views.size > 0)
    console.log("[Info] Created all answer posts (look above)")

  return views
}

/**
 * @param {Post} post
 * @returns {EditorView?}
 */
function generateEmptyPostDOM(post) {
  if (!post) {
    console.log("[Warning] No Feedback Post found !")
    return null
  }

  let view = buildPostDOM(post)
  console.log("[Info] Created the feedback post (look above)")

  return view
}

/**
 * @param {object} flags
 * @param {ViewGroup} views 
 */
function conditionalInitialization(flags, views) {
  enableCollabroationCodeEditor(flags.isEnableCollab, views.unfilled)
  enableCreatingNewPostWithCodeEditor(flags.isCreatePostElement, views.unfilled)

  /**
   * @param {boolean} enabled
   * @param {EditorView?} view
   * @return {void}
   */
  function enableCollabroationCodeEditor(enabled, view) {
    if (!enabled)
      return

    if (!view) {
      console.error("[Error] Unable to build collaboration code editor")
      return
    }

    //
    // Get language value from selection on DOM
    //
    let lang = null

    let postID = getCreationPostID()
    let langElement = getEditorLanguageElement(getPostElement(postID))

    if (langElement) {
      let languageLabel = langElement.value
      lang = getLanguageExtensionFromLabel(languageLabel)
    }

    //
    // Enable users Collaboration on web code editor via WebSocket
    //
    const websocketUrl = `ws://${location.host}/play`

    const ydoc = new Y.Doc()
    const provider = new WebsocketProvider(websocketUrl, 'ws', ydoc)
    provider.awareness.setLocalStateField('user', {
      name: 'user_unknown',
      color: '#335522',
    })

    let sharedContent = ydoc.getText('code-editor')

    /** @type {RemoteCollaboration} */
    let collab = { ydoc, provider, sharedContent }

    let freshState = createEditorState(sharedContent.toString(), lang, collab)
    editEditorView(view, freshState)

    //
    // Register event function that will be called each time a language is swithed to
    //
    if (langElement) {
      /** @param {any} el */
      storage.collaborationEditorEventListenerFunction = function(el) {
        let languageExtension = selectLanguageFromDOM(el.target)

        reconfigureCodeEditor(view, languageExtension, collab)
        console.log("[Info] Successfully switched to Collaboration language : ", el.target.value)
      }

      langElement.removeEventListener("change", storage.standardEditorEventListenerFunction)
      langElement.addEventListener("change", storage.collaborationEditorEventListenerFunction)
    }

    // @ts-ignore
    window.share = { ydoc, provider, sharedContent }
    console.info("[Info] Successfully connected web editor to collaboration network")
  }

  /**
   * @param {boolean} enabled
   * @param {EditorView?} view
   * @returns {void}
   */
  function enableCreatingNewPostWithCodeEditor(enabled, view) {
    if (!enabled)
      return

    if (!view) {
      console.error("[Error] Unable to connect post creation process to backend server")
      return
    }

    let postID = getCreationPostID()
    let postHtmlElement = getPostElement(postID)
    if (!postHtmlElement) {
      console.log("[Error] Unable to find 'Post' element that allow creating new post on page")

      return
    }

    let form = postHtmlElement.parentElement
    if (!form) {
      console.log("[Error] No Parent element found for 'User Post' !")
      console.log("[Info] User wont be able to send/create new Code Snipet Post !!")

      return
    }

    if (form.nodeName != "FORM") {
      console.log("[Error] No form found for 'User Post'. Therefore, can't create new post. Operation aborted")

      return
    }

    let codeEditorLanguageElement = form.querySelector(".code-editor__header select")
    let hiddenCodeInput = form.querySelector("textarea[name=code]")

    if (!codeEditorLanguageElement) {
      console.log("[Error] Unable to find language selector for code editor")
      console.log("[Info] User wont be able to send/create new Code Snipet Post !!")

      return
    }

    if (!hiddenCodeInput) {
      console.log("[Error] Unable to find the PostID through the hidden data field !")
      console.log("[Info] User wont be able to send/create new Code Snipet Post !!")

      return
    }

    // @ts-ignore
    let defaultPostID = codeEditorLanguageElement.dataset.idPost
    if (defaultPostID > 0)
      console.log("[Warning] Default PostID must be 0 or lesser for a post that has yet to be created in the DB. Current PostID = ", defaultPostID)

    form.addEventListener("submit", function(_) {
      let codeContent = view.state.doc
      let codeContentFormated = codeContent.toString()

      // @ts-ignore
      hiddenCodeInput.value = codeContentFormated
    })

  }
}

//
// Init
//

/**
 * @typedef {object} Storage
 * @property {any} standardEditorEventListenerFunction
 * @property {any} collaborationEditorEventListenerFunction
 */

/**
 * @typedef {object} RemoteCollaboration
 * @property {Y.Doc} ydoc
 * @property {WebsocketProvider} provider
 * @property {Y.Text} sharedContent
 */

/**
 * @typedef {object} ViewGroup
 * @property {EditorView?} leader
 * @property {ViewsMap} satellites
 * @property {EditorView?} unfilled
 */

// @ts-ignore
const APP = window.app

/** @type {Storage} */
const storage = {
  standardEditorEventListenerFunction: () => { },
  collaborationEditorEventListenerFunction: () => { },
}

/**
 * @type {ViewGroup}
 */
let VIEWS = {}

if (APP) {
  VIEWS.leader = generateOriginalPostDOM(APP.originalPoster)
  VIEWS.satellites = generateAnswerPostsDOM(APP.answersToPoster)
  VIEWS.unfilled = generateEmptyPostDOM(APP.emptyPost)

  conditionalInitialization(APP.flags, VIEWS)

  // @ts-ignore
  window.app.views = VIEWS
  APP.views = VIEWS
}


