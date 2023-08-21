import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { VerbCard } from './VerbCard'

type Props = {
  module?: Module
}

export const VerbList: React.FC<Props> = ({ module }) => {
  if(!module) return <></>
  const verbs = module?.verbs
  return (
    <>
      {verbs?.map(verb => (
        <div key={verb.verb?.name}>
          <VerbCard verb={verb} />
        </div>
      ))}
    </>
  )
}
