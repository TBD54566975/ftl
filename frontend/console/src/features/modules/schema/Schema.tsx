import { useMemo, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { classNames } from '../../../utils'
import { DeclLink } from '../decls/DeclLink'
import { LinkToken, LinkVerbNameToken } from './LinkTokens'
import { UnderlyingType } from './UnderlyingType'
import { commentPrefix, declTypes, shouldAddLeadingSpace, specialChars, staticKeywords } from './schema.utils'

function maybeRenderDeclName(token: string, declType: string, tokens: string[], i: number, containerRect?: DOMRect) {
  const offset = declType === 'database' ? 4 : 2
  if (i - offset < 0 || declType !== tokens[i - offset]) {
    return
  }
  if (declType === 'enum') {
    return [<LinkToken key='l' token={token} containerRect={containerRect} />]
  }
  if (declType === 'verb') {
    return <LinkVerbNameToken token={token} containerRect={containerRect} />
  }
  return <LinkToken token={token} containerRect={containerRect} />
}

function maybeRenderUnderlyingType(token: string, declType: string, tokens: string[], i: number, moduleName: string, containerRect?: DOMRect) {
  if (declType === 'database') {
    return
  }

  // Parse type(s) out of the headline signature
  const offset = 4
  if (i - offset >= 0 && tokens.slice(0, i - offset + 1).includes(declType)) {
    return <UnderlyingType token={token} containerRect={containerRect} />
  }

  // Parse type(s) out of nested lines
  if (tokens.length > 4 && tokens.slice(0, 4).filter((t) => t !== ' ').length === 0) {
    if (i === 6 && tokens[4] === '+calls') {
      return <UnderlyingType token={token} containerRect={containerRect} />
    }
    if (i === 6 && tokens[4] === '+subscribe') {
      return <DeclLink moduleName={moduleName} declName={token} textColors='font-bold text-green-700 dark:text-green-400' containerRect={containerRect} />
    }
    const plusIndex = tokens.findIndex((t) => t.startsWith('+'))
    if (i >= 6 && (i < plusIndex || plusIndex === -1)) {
      return <UnderlyingType token={token} containerRect={containerRect} />
    }
  }
}

const SchemaLine = ({ line, containerRect }: { line: string; containerRect?: DOMRect }) => {
  const { moduleName } = useParams()
  if (line.startsWith(commentPrefix)) {
    return <span className='text-gray-500 dark:text-gray-400'>{line}</span>
  }
  const tokens = line.split(/( )/).filter((l) => l !== '')
  let declType: string
  return tokens.map((token, i) => {
    if (token.trim() === '') {
      return <span key={i}>{token}</span>
    }
    if (specialChars.includes(token)) {
      return <span key={i}>{token}</span>
    }
    if (staticKeywords.includes(token)) {
      return (
        <span key={i} className='text-fuchsia-700 dark:text-fuchsia-400'>
          {token}
        </span>
      )
    }
    if (declTypes.includes(token) && tokens.length > 2 && tokens[2] !== ' ') {
      declType = token
      return (
        <span key={i} className='text-fuchsia-700 dark:text-fuchsia-400'>
          {token}
        </span>
      )
    }
    if (token[0] === '+' && token.slice(1).match(/^\w+$/)) {
      return (
        <span key={i} className='text-fuchsia-700 dark:text-fuchsia-400'>
          {token}
        </span>
      )
    }

    const numQuotesBefore = (tokens.slice(0, i).join('').match(/"/g) || []).length + (token.match(/^".+/) ? 1 : 0)
    const numQuotesAfter =
      (
        tokens
          .slice(i + 1, tokens.length)
          .join('')
          .match(/"/g) || []
      ).length + (token.match(/.+"$/) ? 1 : 0)
    if (numQuotesBefore % 2 === 1 && numQuotesAfter % 2 === 1) {
      return (
        <span key={i} className='text-rose-700 dark:text-rose-300'>
          {token}
        </span>
      )
    }

    const maybeDeclName = maybeRenderDeclName(token, declType, tokens, i, containerRect)
    if (maybeDeclName) {
      return <span key={i}>{maybeDeclName}</span>
    }
    const maybeUnderlyingType = maybeRenderUnderlyingType(token, declType, tokens, i, moduleName || '', containerRect)
    if (maybeUnderlyingType) {
      return <span key={i}>{maybeUnderlyingType}</span>
    }
    return <span key={i}>{token}</span>
  })
}

export const Schema = ({ schema, containerRect }: { schema: string; containerRect?: DOMRect }) => {
  const ref = useRef<HTMLDivElement>(null)
  const rect = ref?.current?.getBoundingClientRect()
  const ll = useMemo(() => schema.split('\n'), [schema])
  const lines = ll.map((l, i) => (
    <div ref={ref} key={i} className={classNames('mb-1', shouldAddLeadingSpace(ll, i) ? 'mt-4' : '')}>
      <SchemaLine line={l} containerRect={containerRect || rect} />
    </div>
  ))
  return <div className='whitespace-pre font-mono text-xs'>{lines}</div>
}
