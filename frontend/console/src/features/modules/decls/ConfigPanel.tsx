import type { Config } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { Schema } from '../schema/Schema'
import { PanelHeader } from './PanelHeader'
import { References } from './References'

export const ConfigPanel = ({ value, schema, moduleName, declName }: { value: Config; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }
  const decl = value.config
  if (!decl) {
    return
  }
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={false} comments={decl.comments}>
        Config: {moduleName}.{declName}
      </PanelHeader>
      <div className='-mx-3.5'>
        <Schema schema={schema} />
      </div>
      <References references={value.references} />
    </div>
  )
}
