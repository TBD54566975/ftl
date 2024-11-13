import type { Data } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { Schema } from '../schema/Schema'
import { PanelHeader } from './PanelHeader'
import { References } from './References'

export const DataPanel = ({ value, schema, moduleName, declName }: { value: Data; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }
  const decl = value.data
  if (!decl) {
    return
  }
  const maybeTypeParams = decl.typeParameters.length === 0 ? '' : `<${decl.typeParameters.map((p) => p.name).join(', ')}>`
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={decl.export} comments={decl.comments}>
        data: {moduleName}.{declName}
        {maybeTypeParams}
      </PanelHeader>
      <div className='-mx-3.5'>
        <Schema schema={schema} />
      </div>
      <References references={value.references} />
    </div>
  )
}
