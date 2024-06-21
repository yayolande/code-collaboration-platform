
import { lineNumbers, drawSelection, rectangularSelection, highlightActiveLine, crosshairCursor } from '@codemirror/view'
import { keymap, dropCursor, highlightActiveLineGutter, highlightSpecialChars } from '@codemirror/view'
import { syntaxHighlighting, indentOnInput, defaultHighlightStyle, bracketMatching, foldGutter } from '@codemirror/language'
import { autocompletion, closeBrackets } from '@codemirror/autocomplete'
import { highlightSelectionMatches } from '@codemirror/search'
import { defaultKeymap, indentWithTab } from '@codemirror/commands'

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

function createEditorState(codeContent, languageExtension) {
  if (!languageExtension) {
    languageExtension = () => []
  }

  if (!codeContent) {
    codeContent = ""
  }

  codeContent = codeContent.toString().replaceAll(",", "\n")

  let startState = EditorState.create({
    doc: codeContent,
    extensions: [
      lineNumbers(),
      drawSelection(),
      // foldGutter(),
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
      ]),
      languageExtension(),
    ],
  })

  return startState
}

function createEditorView(editorCanvas, startState) {
  editorCanvas.innerHTML = ''

  let view = new EditorView({
    state: startState,
    parent: editorCanvas,
  })

  view.dom.style.minHeight = "300px"

  return view
}

function selectLanguageFromDOM(el) {
  let languageExtension = getLanguageExtensionFromLabel(el.value)

  console.log(`[Info] [setLanguageToCodeEditor] languageExtension = `, languageExtension)

  return languageExtension
}

function getLanguageExtensionFromLabel(languageLabel) {
  let languageExtension = null

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
  }

  return languageExtension
}

function setLanguageToCodeEditor(codeEditorElement, views, postID, languageExtension) {

  if (!codeEditorElement) {
    console.log("[Error] Unable to set language related to a Code Editor since codeEditorElement = null")
    return
  }

  let previousView = views[postID]

  let startState = createEditorState(previousView.state.doc, languageExtension)
  let view = createEditorView(codeEditorElement, startState)

  views[postID] = view
}

function getPostElement(postID) {
  return document.getElementById(`user-post__${postID}`)
}

function getCreationPostID() {
  let defaultPostID = 0
  let post = GLOBAL_DATA_EMPTY_POSTER

  if (!post) {
    console.log("[Warning] No Data for empty 'Post' found")
    console.log("[Info] Post Creation may not work as expected !")

    return defaultPostID
  }

  return post.PostID
}

function generateRawPostOnDOM(post, views) {
  let postID = post.PostID

  let postHtmlElement = getPostElement(postID)
  if (!postHtmlElement) {
    console.log("[Warning - generateRawPostOnDOM()] postHtmlElement == null, abording Code Editor binding ...")
    return views
  }

  let codeEditorElement = postHtmlElement.getElementsByClassName("code-editor__body")[0]
  let codeEditorLanguageElement = postHtmlElement.querySelector(".code-editor__header select")

  if (!codeEditorElement || !codeEditorLanguageElement) {
    console.log("[Warning - generateRawPostOnDOM()] codeEditor(Language)Element == null, abording Code Editor binding ...")
    return views
  }

  let posterCode = post.CodeSnipet

  let startState = createEditorState(posterCode, null)
  let view = createEditorView(codeEditorElement, startState)

  views[postID] = view

  let languageLabel = post.LanguageCode
  if (!languageLabel) {
    languageLabel = "js"

    console.log("[Warning] Language for code editor was empty ! Value set to default : ", languageLabel)
  }

  let languageExtension = getLanguageExtensionFromLabel(languageLabel)
  setLanguageToCodeEditor(codeEditorElement, views, postID, languageExtension)

  codeEditorLanguageElement.value = languageLabel
  codeEditorLanguageElement.addEventListener("change", function(el) {
    let postID = el.target.dataset.idPost

    let languageExtension = selectLanguageFromDOM(el.target)
    let codeEditorElement = getPostElement(postID).getElementsByClassName("code-editor__body")[0]

    setLanguageToCodeEditor(codeEditorElement, views, postID, languageExtension)
  })

  return views
}

function generateOriginalPostOnDOM(post, views) {

  if (!post) {
    console.log("[Error] No Original Post Found !")
    return views
  }

  if (!views) {
    console.log("[Warning] Global Code Editor 'VIEWS' is null")
    views = {}
  }

  views = generateRawPostOnDOM(post, views)
  console.log("[Info] Found Original Post :")

  return views
}

function generateAnswerPostOnDOM(posts, views) {

  if (!posts) {
    console.log("[Warning] No Answer Posts Found !")
    return views
  }

  if (!views) {
    console.log("[Warning] Global Code Editor 'VIEWS' is null")
    views = {}
  }

  for (let i = 0; i < posts.length; i++) {
    views = generateRawPostOnDOM(posts[i], views)
    console.log("[Info] Found a post :")
  }

  return views
}

function generateEmptyPostOnDom(post, views) {

  if (!post) {
    console.log("[Warning] No Creation Post found !")
    return views
  }

  if (!views) {
    console.log("[Warning] Global Code Editor 'VIEWS' is null")
    views = {}
  }

  views = generateRawPostOnDOM(post, views)

  return views
}

function conditionalInitialization(flags, views) {
  function registerHookForPostCreation(views, postID) {

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
    let hiddenCodeInput = form.querySelector("input[name=code]")

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

    // INFO: Why not 'postID' ?
    let defaultPostID = codeEditorLanguageElement.dataset.idPost
    if (defaultPostID > 0)
      console.log("[Warning] Default PostID must be 0 or lesser for a post that has yet to be created in the DB. Current PostID = ", defaultPostID)

    form.addEventListener("submit", function(_) {
      let codeContent = views[defaultPostID].state.doc
      let codeContentFormated = codeContent.text.toString()

      hiddenCodeInput.value = codeContentFormated
    })

  }

  if (flags.isCreatePostElement) {
    let postID = getCreationPostID()
    registerHookForPostCreation(views, postID)
  }
}

//
// Init
//

let VIEWS = {}

VIEWS = generateOriginalPostOnDOM(GLOBAL_DATA_ORIGINAL_POSTER, VIEWS)
VIEWS = generateAnswerPostOnDOM(GLOBAL_DATA_ANSWERS_POSTER, VIEWS)
VIEWS = generateEmptyPostOnDom(GLOBAL_DATA_EMPTY_POSTER, VIEWS)

GLOBAL_VIEWS = VIEWS

const FLAGS = GLOBAL_FLAGS
conditionalInitialization(FLAGS, VIEWS)

