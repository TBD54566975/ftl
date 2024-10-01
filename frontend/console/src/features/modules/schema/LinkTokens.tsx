import { useParams } from 'react-router-dom'
import { DeclLink } from '../decls/DeclLink'
import { UnderlyingType } from './UnderlyingType'

export const LinkToken = ({ token, containerRect }: { token: string; containerRect?: DOMRect }) => {
  const { moduleName } = useParams()
  if (token.match(/^\w+$/)) {
    return (
      <span className='font-bold'>
        <DeclLink slim moduleName={moduleName} declName={token} containerRect={containerRect} />
      </span>
    )
  }
  if (token.match(/^\w+<\w+/)) {
    const splitToken = token.split('<')
    return (
      <>
        <span className='font-bold'>
          <DeclLink slim moduleName={moduleName} declName={splitToken[0]} containerRect={containerRect} />
        </span>
        <span className='text-green-700 dark:text-green-400'>
          {'<'}
          <UnderlyingType token={splitToken[1]} containerRect={containerRect} />
        </span>
      </>
    )
  }
  return token
}

export const LinkVerbNameToken = ({ token, containerRect }: { token: string; containerRect?: DOMRect }) => {
  const splitToken = token.split('(')
  if (splitToken.length < 2) {
    return
  }
  return (
    <span>
      <LinkToken token={splitToken[0]} containerRect={containerRect} />
      (<UnderlyingType token={splitToken[1]} containerRect={containerRect} />
    </span>
  )
}
