import {useEffect, useState} from 'react'

export interface InitialState {
  initialText: string
  schema?: string
  readonly?: boolean
}

export const CodeEditor = (
  { initialState, onTextChanged }:
    { initialState: InitialState, onTextChanged?: (text: string) => void }
) => {
  const [editorText, setEditorText] = useState(initialState.initialText)

  useEffect(() => {
    setEditorText(initialState.initialText)
  }, [initialState])

  const handleChange = (text: string) => {
    setEditorText(text)
    if (onTextChanged) {
      onTextChanged(text)
    }
  }

  return (
    <textarea
      value={editorText}
      readOnly={initialState.readonly}
      onChange={(e) => handleChange(e.target.value)}
      style={{ width: '100%',fontFamily: 'monospace' }}
    />
  )
}
