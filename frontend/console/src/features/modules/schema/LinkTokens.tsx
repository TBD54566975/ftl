import { useParams } from 'react-router-dom'
import { DeclLink } from '../decls/DeclLink'
import { UnderlyingType } from './UnderlyingType'

export const LinkToken = ({ token }: { token: string }) => {
  const { moduleName } = useParams()
  if (token.match(/^\w+$/)) {
    return (
      <span className='font-bold'>
        <DeclLink slim moduleName={moduleName} declName={token} />
      </span>
    )
  }
  return token
}

export const LinkVerbNameToken = ({ token }: { token: string }) => {
  const splitToken = token.split('(')
  return (
    <span>
      <LinkToken token={splitToken[0]} />
      (<UnderlyingType token={splitToken[1]} />
    </span>
  )
}
