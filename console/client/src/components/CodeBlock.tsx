import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { atomDark, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { useDarkMode } from '../providers/dark-mode-provider'

interface Props {
  code: string
  language: string
  maxHeight?: number
}

export const CodeBlock = ({ code, language, maxHeight = 300 }: Props) => {
  const { isDarkMode } = useDarkMode()

  return (
    <SyntaxHighlighter
      wrapLongLines={true}
      lineProps={{ style: { flexWrap: 'wrap' } }}
      language={language}
      style={isDarkMode ? atomDark : oneLight}
      customStyle={{
        fontSize: '12px',
        maxHeight: `${maxHeight}px`,
        overflow: 'auto',
        maxWidth: `100%`,
      }}
    >
      {code}
    </SyntaxHighlighter>
  )
}
