import { Module, Verb } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { VerbCard } from './VerbCard'

type Props = {
  module?: Module
}

export const VerbList: React.FC<Props> = ({ module }) => {
  const verbs = module?.decls.filter(decl => decl.value.case === 'verb')

  return (
    <>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3 py-6">
        {verbs?.map(verb => (
          <VerbCard key={verb.value.value?.name} module={module} verb={verb.value.value as Verb} />
        ))}
      </div>
    </>
  )
}
