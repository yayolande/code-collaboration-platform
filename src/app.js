
import { lineNumbers, drawSelection, rectangularSelection, highlightActiveLine, crosshairCursor } from '@codemirror/view'
import { keymap, dropCursor, highlightActiveLineGutter, highlightSpecialChars } from '@codemirror/view'
import { syntaxHighlighting, indentOnInput, defaultHighlightStyle, bracketMatching, foldGutter } from '@codemirror/language'
import { autocompletion, closeBrackets } from '@codemirror/autocomplete'
import { highlightSelectionMatches } from '@codemirror/search'
import { defaultKeymap, indentWithTab } from '@codemirror/commands'
import { javascript } from '@codemirror/lang-javascript'

import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
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
    codeContent = "// Hello World !\n console.log('Melissa')"
  }

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

//
// First, hook the language selection dropdown to code editor 
//
let codeEditorWrapper = document.getElementsByClassName("code-editor")[0]
let codeEditorElement = codeEditorWrapper.getElementsByClassName("code-editor__body")[0]
let codeEditorLanguageElement = codeEditorWrapper.querySelector(".code-editor__header > select")
let defaultLanguageExtension = () => [javascript()]

console.log(`${codeEditorLanguageElement}`)
codeEditorLanguageElement.addEventListener("change", (el) => {
  let languageExtension

  switch (el.target.value) {
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

  startState = createEditorState(view.state.doc, languageExtension)
  view = createEditorView(codeEditorElement, startState)
})

let startState = createEditorState(null, defaultLanguageExtension)
let view = createEditorView(codeEditorElement, startState)

console.log(`${startState.doc}`)
