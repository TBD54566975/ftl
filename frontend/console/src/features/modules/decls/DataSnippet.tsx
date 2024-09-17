import type { Data } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { TypeEl } from './TypeEl'

export const DataSnippet = ({ value }: { value: Data }) => {
  const maybeTypeParams = value.typeParameters.length === 0 ? '' : `<${value.typeParameters.map((p) => p.name).join(', ')}>`
  return (
    <div className='text-xs font-mono'>
      <div>
        {value.export ? 'export ' : ''}
        data {value.name}
        {maybeTypeParams}
        {value.fields.length === 0 ? ' {}' : ' {'}
      </div>
      {value.fields.length === 0 || (
        <div className='text-xs font-mono inline-grid grid-cols-2 gap-x-4 gap-y-2 ml-8 my-2' style={{ gridTemplateColumns: 'auto auto' }}>
          {value.fields.map((f, i) => [<span key={`field-name-${i}`}>{f.name}</span>, <TypeEl key={`field-type-${i}`} t={f.type} />])}
        </div>
      )}
      <div>{value.fields.length === 0 ? '' : '}'}</div>
    </div>
  )
}
