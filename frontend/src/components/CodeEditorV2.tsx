import { EditorState, type Extension } from '@codemirror/state'
import { EditorView, drawSelection, gutter, highlightActiveLineGutter, hoverTooltip, keymap, lineNumbers } from '@codemirror/view'

import { bracketMatching, defaultHighlightStyle, foldGutter, foldKeymap, indentOnInput, syntaxHighlighting } from '@codemirror/language'
import { lintGutter } from '@codemirror/lint'
import { lintKeymap } from '@codemirror/lint'
import { linter } from '@codemirror/lint'

import { autocompletion, closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete'
import { atomone } from '@uiw/codemirror-theme-atomone'
import { githubLight } from '@uiw/codemirror-theme-github'

import { defaultKeymap, history, indentWithTab, redo, undo } from '@codemirror/commands'
import { handleRefresh, jsonSchemaHover, jsonSchemaLinter, stateExtensions } from 'codemirror-json-schema'
import { json5, json5ParseLinter } from 'codemirror-json5'
import { useEffect, useRef, useState } from 'react'
import { useUserPreferences } from '../providers/user-preferences-provider'

export const CodeEditor = ({ content, schema, readOnly, onChange, defaultContent }: {
  content: string,
  schema?: string,
  readOnly?: boolean,
  onChange?: (text: string) => void,
  defaultContent?: string
}) => {
  const { isDarkMode } = useUserPreferences()
  const containerRef = useRef(null)
  const viewRef = useRef<EditorView | null>(null)
  const [lastReset, setLastReset] = useState(Date.now())

  useEffect(() => {
    const sch = schema ? JSON.parse(schema) : null

    const editingExtensions: Extension[] =
      readOnly || false
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
        stateExtensions(sch),
      ]

    viewRef.current = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions: [
          ...editingExtensions,
          EditorView.updateListener.of(({ state }) => {
            if (onChange) {
              onChange(state.doc.toString())
            }
          }),
          EditorView.theme({
            '&': { height: '100%' },
          }),
          isDarkMode ? atomone : githubLight,
          json5(),
          gutter({ class: 'CodeMirror-lint-markers' }),
          highlightActiveLineGutter(),
          closeBrackets(),
          EditorView.lineWrapping,
          EditorState.tabSize.of(2),
          history(),
          keymap.of([
            ...defaultKeymap,
            ...closeBracketsKeymap,
            ...foldKeymap,
            ...lintKeymap,
            indentWithTab,
            { key: "Mod-z", run: undo, preventDefault: true },
            { key: "Mod-Shift-z", run: redo, preventDefault: true },
          ]),
          bracketMatching(),
          syntaxHighlighting(defaultHighlightStyle)
        ],
      }),
      parent: containerRef.current || undefined,
    })

    return () => {
      viewRef.current?.destroy()
      viewRef.current = null
    }
  }, [lastReset, schema, readOnly, isDarkMode])

  useEffect(() => {
    if (viewRef.current && viewRef.current.state.doc.toString() !== content) {
      viewRef.current.dispatch({
        changes: { from: 0, to: viewRef.current.state.doc.length, insert: "" }
      });
    }
  }, [content])

  function onReset() {
    if (!onChange || !defaultContent) return
    onChange(defaultContent)
    setLastReset(Date.now())
  }

  return (
    <div className='h-full'>
      {defaultContent && <div className='absolute z-10 top-2 text-sm px-2 py-0.5 right-4 cursor-pointer shadow-lg rounded-md bg-gray-300 hover:bg-gray-200 dark:bg-gray-900 hover:dark:bg-gray-700' onClick={onReset}>Reset</div>}
      <div className='h-full' ref={containerRef} />
    </div>
  )
}
