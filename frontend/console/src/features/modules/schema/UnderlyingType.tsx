import { DeclLink } from '../decls/DeclLink'

export const UnderlyingType = ({ token }: { token: string }) => {
  if (token.match(/^\[.+\]$/)) {
    // Handles lists: [elementType]
    return (
      <span className='text-green-700 dark:text-green-400'>
        [<UnderlyingType token={token.slice(1, token.length - 1)} />]
      </span>
    )
  }

  if (token.match(/^{.+:$/)) {
    // Handles first token of map: {KeyType: ValueType}
    return (
      <span className='text-green-700 dark:text-green-400'>
        {'{'}
        <UnderlyingType token={token.slice(1, token.length - 1)} />:
      </span>
    )
  }

  if (token.match(/.+}$/)) {
    // Handles last token of map: {KeyType: ValueType}
    return (
      <span className='text-green-700 dark:text-green-400'>
        <UnderlyingType token={token.slice(0, token.length - 1)} />
        {'}'}
      </span>
    )
  }

  if (token.match(/^.+\?$/)) {
    // Handles optional: elementType?
    return (
      <span className='text-green-700 dark:text-green-400'>
        <UnderlyingType token={token.slice(0, token.length - 1)} />?
      </span>
    )
  }

  if (token.match(/^.+\)$/)) {
    // Handles closing parens in param list of verb signature: verb echo(inputType) outputType
    return (
      <span>
        <UnderlyingType token={token.slice(0, token.length - 1)} />)
      </span>
    )
  }

  const maybeSplitRef = token.split('.')
  if (maybeSplitRef.length < 2) {
    // Not linkable because it's not a ref
    return <span className='text-green-700 dark:text-green-400'>{token}</span>
  }
  const moduleName = maybeSplitRef[0]
  const declName = maybeSplitRef[1].split('<')[0]
  const primaryTypeEl = (
    <span className='text-green-700 dark:text-green-400'>
      <DeclLink moduleName={moduleName} declName={declName.split(/[,>]/)[0]} textColors='font-bold text-green-700 dark:text-green-400' />
      {[',', '>'].includes(declName.slice(-1)) ? declName.slice(-1) : ''}
    </span>
  )
  const hasTypeParams = maybeSplitRef[1].includes('<')
  if (!hasTypeParams) {
    return primaryTypeEl
  }
  return (
    <span className='text-green-700 dark:text-green-400'>
      {primaryTypeEl}
      {'<'}
      <UnderlyingType token={maybeSplitRef.length === 2 ? maybeSplitRef[1].split('<')[1] : `${maybeSplitRef[1].split('<')[1]}.${maybeSplitRef.slice(2)}`} />
    </span>
  )
}
