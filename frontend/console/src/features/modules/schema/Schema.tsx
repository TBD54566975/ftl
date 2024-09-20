import { useParams } from 'react-router-dom'
import { classNames } from '../../../utils'
import { DeclLink } from '../decls/DeclLink'
import { LinkToken, LinkVerbNameToken } from './LinkTokens'
import { UnderlyingType } from './UnderlyingType'
import { commentPrefix, declTypes, isFirstLineOfBlock, specialChars, staticKeywords } from './schema.utils'

function maybeRenderDeclName(token: string, declType: string, tokens: string[], i: number) {
  const offset = declType === 'database' ? 4 : 2
  if (i - offset < 0 || declType !== tokens[i - offset]) {
    return
  }
  if (declType === 'enum') {
    return [<LinkToken key='l' token={token.slice(0, token.length - 1)} />, token.slice(-1)]
  }
  if (declType === 'verb') {
    return <LinkVerbNameToken token={token} />
  }
  return <LinkToken token={token} />
}

function maybeRenderUnderlyingType(token: string, declType: string, tokens: string[], i: number, moduleName: string) {
  if (declType === 'database') {
    return
  }

  // Parse type(s) out of the headline signature
  const offset = 4
  if (i - offset >= 0 && tokens.slice(0, i - offset + 1).includes(declType)) {
    return <UnderlyingType token={token} />
  }

  // Parse type(s) out of nested lines
  if (tokens.length > 4 && tokens.slice(0, 4).filter((t) => t !== ' ').length === 0) {
    if (i === 6 && tokens[4] === '+calls') {
      return <UnderlyingType token={token} />
    }
    if (i === 6 && tokens[4] === '+subscribe') {
      return <DeclLink moduleName={moduleName} declName={token} textColors='font-bold text-green-700 dark:text-green-400' />
    }
    const plusIndex = tokens.findIndex((t) => t.startsWith('+'))
    if (i >= 6 && (i < plusIndex || plusIndex === -1)) {
      return <UnderlyingType token={token} />
    }
  }
}

const SchemaLine = ({ line }: { line: string }) => {
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

    const maybeDeclName = maybeRenderDeclName(token, declType, tokens, i)
    if (maybeDeclName) {
      return <span key={i}>{maybeDeclName}</span>
    }
    const maybeUnderlyingType = maybeRenderUnderlyingType(token, declType, tokens, i, moduleName || '')
    if (maybeUnderlyingType) {
      return <span key={i}>{maybeUnderlyingType}</span>
    }
    return <span key={i}>{token}</span>
  })
}

export const Schema = ({ schema }: { schema: string }) => {
  const ll = schema.split('\n')
  const lines = ll.map((l, i) => (
    <div key={i} className={classNames('mb-1', isFirstLineOfBlock(ll, i) ? 'mt-4' : '')}>
      <SchemaLine line={l} />
    </div>
  ))
  return <div className='mt-4 mx-4 whitespace-pre font-mono text-xs'>{lines}</div>
}
