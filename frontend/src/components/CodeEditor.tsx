import { EditorState, type Extension } from '@codemirror/state'
import { EditorView, drawSelection, gutter, highlightActiveLineGutter, hoverTooltip, keymap, lineNumbers } from '@codemirror/view'

import { bracketMatching, foldGutter, foldKeymap, indentOnInput } from '@codemirror/language'
import { lintGutter } from '@codemirror/lint'
import { lintKeymap } from '@codemirror/lint'
import { linter } from '@codemirror/lint'

import { autocompletion, closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete'
import { atomone } from '@uiw/codemirror-theme-atomone'
import { githubLight } from '@uiw/codemirror-theme-github'

import { defaultKeymap } from '@codemirror/commands'
import { handleRefresh, jsonSchemaHover, jsonSchemaLinter, stateExtensions } from 'codemirror-json-schema'
import { json5, json5ParseLinter } from 'codemirror-json5'
import { useCallback, useEffect, useRef } from 'react'
import { useDarkMode } from '../hooks/use-dark-mode'

const commonExtensions = [
  gutter({ class: 'CodeMirror-lint-markers' }),
  bracketMatching(),
  highlightActiveLineGutter(),
  closeBrackets(),
  keymap.of([...closeBracketsKeymap, ...foldKeymap, ...lintKeymap, ...defaultKeymap]),
  EditorView.lineWrapping,
  EditorState.tabSize.of(2),
]

export interface InitialState {
  initialText: string
  schema?: string
  readonly?: boolean
}

export const CodeEditor = ({ initialState, onTextChanged }: { initialState: InitialState; onTextChanged?: (text: string) => void }) => {
  const { isDarkMode } = useDarkMode()
  const editorContainerRef = useRef(null)
  const editorViewRef = useRef<EditorView | null>(null)

  const handleEditorTextChange = useCallback(
    (state: EditorState) => {
      const currentText = state.doc.toString()
      if (onTextChanged) {
        onTextChanged(currentText)
      }
    },
    [onTextChanged],
  )

  useEffect(() => {
    if (editorContainerRef.current) {
      const sch = initialState.schema ? JSON.parse(initialState.schema) : null

      const editingExtensions: Extension[] =
        initialState.readonly || false
          ? [EditorState.readOnly.of(true)]
          : [
              autocompletion(),
              lineNumbers(),
              lintGutter(),
              indentOnInput(),
              drawSelection(),
              foldGutter(),
              linter(json5ParseLinter(), {
                delay: 300,
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
        extensions: [commonExtensions, isDarkMode ? atomone : githubLight, json5(), editingExtensions],
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
