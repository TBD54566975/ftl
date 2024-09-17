import type { TypeAlias } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { TypeEl } from './TypeEl'

export const TypeAliasSnippet = ({ value }: { value: TypeAlias }) => {
  return (
    <div className='font-mono text-xs'>
      {value.export ? 'export ' : ''}
      typealias {value.name} <TypeEl t={value.type} />
    </div>
  )
}
