import type { Data } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { PanelHeader } from './PanelHeader'
import { TypeEl } from './TypeEl'

export const DataPanel = ({ value, moduleName, declName }: { value: Data; moduleName: string; declName: string }) => {
  const maybeTypeParams = value.typeParameters.length === 0 ? '' : `<${value.typeParameters.map((p) => p.name).join(', ')}>`
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={value.export} comments={value.comments}>
        <p>
          data: {moduleName}.{declName}
          {maybeTypeParams}
        </p>
      </PanelHeader>
      {value.fields.length === 0 || <div className='mt-8 mb-3'>Fields</div>}
      <div className='text-xs font-mono inline-grid grid-cols-2 gap-x-4 gap-y-2' style={{ gridTemplateColumns: 'auto auto' }}>
        {value.fields.map((f, i) => [
          <span key={`field-name-${i}`}>{f.name}</span>,
          <span key={`field-type-${i}`}>
            <TypeEl t={f.type} />
          </span>,
        ])}
      </div>
    </div>
  )
}
