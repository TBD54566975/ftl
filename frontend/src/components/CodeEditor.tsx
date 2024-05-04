import { EditorState, Extension } from '@codemirror/state'
import {
  gutter,
  EditorView,
  hoverTooltip,
  lineNumbers,
  drawSelection,
  keymap,
  highlightActiveLineGutter,
} from '@codemirror/view'

import { lintGutter } from '@codemirror/lint'
import { lintKeymap } from '@codemirror/lint'
import { linter } from '@codemirror/lint'
import {

  indentOnInput,
  bracketMatching,
  foldGutter,
  foldKeymap,
} from '@codemirror/language'

import { githubLight } from '@uiw/codemirror-theme-github'
import { atomone } from '@uiw/codemirror-theme-atomone'
import {
  autocompletion,
  closeBrackets,
  closeBracketsKeymap,
} from '@codemirror/autocomplete'

import { useRef, useEffect, useCallback } from 'react'
import { json, jsonParseLinter } from '@codemirror/lang-json'
import { jsonSchemaLinter, jsonSchemaHover, stateExtensions, handleRefresh } from 'codemirror-json-schema'
import { useDarkMode } from '../providers/dark-mode-provider'
import { defaultKeymap } from '@codemirror/commands'

const commonExtensions = [
  gutter({ class: 'CodeMirror-lint-markers' }),
  bracketMatching(),
  highlightActiveLineGutter(),
  closeBrackets(),
  keymap.of([
    ...closeBracketsKeymap,
    ...foldKeymap,
    ...lintKeymap,
    ...defaultKeymap
  ]),
  EditorView.lineWrapping,
  EditorState.tabSize.of(2),
]

export interface InitialState {
  initialText: string
  schema?: string
  readonly?: boolean
}

export const CodeEditor = (
  { initialState, onTextChanged }:
    { initialState: InitialState, onTextChanged?: (text: string) => void }
) => {
  const { isDarkMode } = useDarkMode()
  const editorContainerRef = useRef(null)
  const editorViewRef = useRef<EditorView | null>(null)

  const handleEditorTextChange = useCallback((state: EditorState) => {
    const currentText = state.doc.toString()
    onTextChanged && onTextChanged(currentText)
  }, [onTextChanged])

  useEffect(() => {
    if (editorContainerRef.current) {
      const sch = initialState.schema ? JSON.parse(initialState.schema) : null

      const editingExtensions: Extension[] = initialState.readonly || false ? [
        EditorState.readOnly.of(true)
      ] : [
        autocompletion(),
        lineNumbers(),
        lintGutter(),
        indentOnInput(),
        drawSelection(),
        foldGutter(),
        linter(jsonParseLinter(), {
          delay: 300
        }),
        linter(jsonSchemaLinter(), {
          needsRefresh: handleRefresh,
        }),
        hoverTooltip(jsonSchemaHover()),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            handleEditorTextChange(update.state)
          }
        }),
        stateExtensions(sch),
      ]

      const state = EditorState.create({
        doc: initialState.initialText,
        extensions: [
          commonExtensions,
          isDarkMode ? atomone : githubLight,
          json(),

          editingExtensions
        ],
      })

      const view = new EditorView({
        state,
        parent: editorContainerRef.current,
      })

      editorViewRef.current = view

      return () => {
        view.destroy()
      }
    }
  }, [initialState, isDarkMode])

  return <div ref={editorContainerRef} />
}
