import { EditorState, type Extension } from '@codemirror/state'
import { EditorView, drawSelection, gutter, highlightActiveLineGutter, hoverTooltip, keymap, lineNumbers } from '@codemirror/view'

import { bracketMatching, foldGutter, foldKeymap, indentOnInput } from '@codemirror/language'
import { lintGutter } from '@codemirror/lint'
import { lintKeymap } from '@codemirror/lint'
import { linter } from '@codemirror/lint'

import { autocompletion, closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete'
import { atomone } from '@uiw/codemirror-theme-atomone'
import { githubLight } from '@uiw/codemirror-theme-github'

import { defaultKeymap, indentWithTab } from '@codemirror/commands'
import { handleRefresh, jsonSchemaHover, jsonSchemaLinter, stateExtensions } from 'codemirror-json-schema'
import { json5, json5ParseLinter } from 'codemirror-json5'
import { useEffect, useRef } from 'react'
import { useUserPreferences } from '../providers/user-preferences-provider'

const commonExtensions = [
  gutter({ class: 'CodeMirror-lint-markers' }),
  bracketMatching(),
  highlightActiveLineGutter(),
  closeBrackets(),
  keymap.of([indentWithTab, ...closeBracketsKeymap, ...foldKeymap, ...lintKeymap, ...defaultKeymap]),
  EditorView.lineWrapping,
  EditorState.tabSize.of(2),
]

export const CodeEditor = ({
  value = '',
  onTextChanged,
  readonly = false,
  schema,
  id,
}: {
  value: string
  onTextChanged?: (text: string) => void
  readonly?: boolean
  schema?: string
  id?: string
}) => {
  const { isDarkMode } = useUserPreferences()
  const editorContainerRef = useRef(null)
  const editorViewRef = useRef<EditorView | null>(null)

  useEffect(() => {
    if (editorContainerRef.current) {
      const sch = schema ? JSON.parse(schema) : null

      const editingExtensions: Extension[] = readonly
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
                const currentText = update.state.doc.toString()
                if (onTextChanged) {
                  onTextChanged(currentText)
                }
              }
            }),
            stateExtensions(sch),
          ]

      const state = EditorState.create({
        doc: value,
        extensions: [
          ...commonExtensions,
          isDarkMode ? atomone : githubLight,
          json5(),
          editingExtensions,
          EditorView.theme({
            '&': { height: '100%' },
          }),
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
  }, [isDarkMode, readonly, schema])

  useEffect(() => {
    if (editorViewRef.current && value !== undefined) {
      const currentText = editorViewRef.current.state.doc.toString()
      if (currentText !== value) {
        const { state } = editorViewRef.current
        const transaction = state.update({
          changes: { from: 0, to: state.doc.length, insert: value },
        })
        editorViewRef.current.dispatch(transaction)
      }
    }
  }, [value])

  return <div id={id} className='h-full' ref={editorContainerRef} />
}
