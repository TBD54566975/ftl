import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { VerbCard } from './VerbCard'

interface Props {
  module?: Module
}

export const VerbList = ({ module }: Props) => {
  const verbs = module?.verbs

  return (
    <>
      <div className='grid grid-cols-1 gap-4 sm:grid-cols-3 py-6'>
        {verbs?.map((verb) => (
          <VerbCard key={verb.verb?.name} module={module} verb={verb} />
        ))}
      </div>
    </>
  )
}
