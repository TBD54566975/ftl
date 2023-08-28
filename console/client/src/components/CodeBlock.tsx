import React from 'react'
import {Prism as SyntaxHighlighter} from 'react-syntax-highlighter'
import {
  atomDark,
  oneLight,
} from 'react-syntax-highlighter/dist/esm/styles/prism'
import {useDarkMode} from '../providers/dark-mode-provider'

type Props = {
  code: string
  language: string
}

export const CodeBlock: React.FC<Props> = ({code, language}) => {
  const {isDarkMode} = useDarkMode()

  return (
    <SyntaxHighlighter
      language={language}
      style={
        isDarkMode
          ? (atomDark as React.CSSProperties)
          : (oneLight as React.CSSProperties)
      }
      customStyle={{fontSize: '12px'}}
    >
      {code}
    </SyntaxHighlighter>
  )
}
