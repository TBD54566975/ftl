import { Module, Verb } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { VerbRow } from './VerbRow'

type Props = {
  module?: Module
}

export const VerbList: React.FC<Props> = ({ module }) => {
  const verbs = module?.decls.filter(decl => decl.value.case === 'verb')

  return (
    <>
      <dl role="list" className="divide-y divide-black/5 dark:divide-white/5">
        {verbs?.map(verb => (
          <VerbRow
            key={verb.value.value?.name}
            verb={verb.value.value as Verb}
          />
        ))}
      </dl>
    </>
  )
}
