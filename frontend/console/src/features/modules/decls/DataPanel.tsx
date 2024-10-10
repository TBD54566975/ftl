import type { Data } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { Schema } from '../schema/Schema'
import { PanelHeader } from './PanelHeader'

export const DataPanel = ({ value, schema, moduleName, declName }: { value: Data; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }
  const maybeTypeParams = value.typeParameters.length === 0 ? '' : `<${value.typeParameters.map((p) => p.name).join(', ')}>`
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={value.export} comments={value.comments}>
        data: {moduleName}.{declName}
        {maybeTypeParams}
      </PanelHeader>
      <div className='-mx-3.5'>
        <Schema schema={schema} />
      </div>
    </div>
  )
}
