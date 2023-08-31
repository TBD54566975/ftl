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
  maxHeight?: number
}

export const CodeBlock: React.FC<Props> = ({
  code,
  language,
  maxHeight = 300,
}) => {
  const {isDarkMode} = useDarkMode()

  return (
    <SyntaxHighlighter
      wrapLongLines={true}
      lineProps={{style: {flexWrap: 'wrap'}}}
      language={language}
      style={
        isDarkMode
          ? (atomDark as React.CSSProperties)
          : (oneLight as React.CSSProperties)
      }
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
