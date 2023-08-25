import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { atomDark, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { useDarkMode } from '../providers/dark-mode-provider'

interface Props {
  code: string
  language: string
}

export const CodeBlock = ({ code, language }: Props) => {
  const { isDarkMode } = useDarkMode()

  return (
    <SyntaxHighlighter language={language} style={isDarkMode ? atomDark : oneLight} customStyle={{ fontSize: '12px' }}>
      {code}
    </SyntaxHighlighter>
  )
}
