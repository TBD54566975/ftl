import {useEffect, useState} from 'react'
import {useDarkMode} from '../providers/dark-mode-provider.tsx'

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
  const { isDarkMode } = useDarkMode()

  useEffect(() => {
    setEditorText(initialState.initialText)
  }, [initialState])

  const handleChange = (text: string) => {
    setEditorText(text)
    if (onTextChanged) {
      onTextChanged(text)
    }
  }

  const extraClass = (isDarkMode) ? 'bg-gray-800 text-white' : 'bg-gray-100 text-black'

  return (
    <textarea
      value={editorText}
      readOnly={initialState.readonly}
      onChange={(e) => handleChange(e.target.value)}
      className={`w-full h-full p-2 text-sm font-mono ${extraClass}`}
    />
  )
}
