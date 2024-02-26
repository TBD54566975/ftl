import hljs from 'highlight.js/lib/core'
import go from 'highlight.js/lib/languages/go'
import graphql from 'highlight.js/lib/languages/graphql'
import json from 'highlight.js/lib/languages/json'
import plaintext from 'highlight.js/lib/languages/plaintext'
import 'highlight.js/styles/atom-one-dark.css'
import { useEffect, useRef } from 'react'

interface Props {
  code: string
  language: string
  maxHeight?: number
}

export const CodeBlock = ({ code, language, maxHeight }: Props) => {
  const codeRef = useRef<HTMLElement>(null)

  useEffect(() => {
    hljs.configure({ ignoreUnescapedHTML: true })
    hljs.registerLanguage('graphql', graphql)
    hljs.registerLanguage('json', json)
    hljs.registerLanguage('go', go)
    hljs.registerLanguage('plaintext', plaintext)

    if (codeRef.current) {
      codeRef.current.removeAttribute('data-highlighted')
      hljs.highlightElement(codeRef.current)
    }

    return () => {
      if (codeRef.current) {
        codeRef.current.removeAttribute('data-highlighted')
      }
    }
  }, [code, language])

  return (
    <pre style={{ maxHeight: maxHeight ? `${maxHeight}px` : 'auto' }}>
      <code ref={codeRef} className={`language-${language} text-xs`}>
        {code}
      </code>
    </pre>
  )
}
