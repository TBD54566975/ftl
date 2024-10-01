import type { Enum } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { Schema } from '../schema/Schema'
import { PanelHeader } from './PanelHeader'
import { enumType } from './enum.utils'

export const EnumPanel = ({ value, schema, moduleName, declName }: { value: Enum; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={value.export} comments={value.comments}>
        {enumType(value)} Enum: {moduleName}.{declName}
      </PanelHeader>
      <div className='-mx-3.5'>
        <Schema schema={schema} />
      </div>
    </div>
  )
}
