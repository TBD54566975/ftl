import { Data, Module } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import React from 'react'
import { TypeRow } from './TypeRow'

type Props = {
  module?: Module
}

export const TypeList: React.FC<Props> = ({ module }) => {
  const types = module?.decls.filter(decl => decl.value.case === 'data')
  return (
    <>
      <dl role="list" className="divide-y divide-black/5 dark:divide-white/5">
        {types?.map(type => (
          <TypeRow key={type.value.value?.name} data={type.value.value as Data} />
        ))}
      </dl>
    </>
  )
}
