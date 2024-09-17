import type { Enum } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { EnumSnippet } from './EnumSnippet'
import { PanelHeader } from './PanelHeader'
import { enumType } from './enum.utils'

export const EnumPanel = ({ value, moduleName, declName }: { value: Enum; moduleName: string; declName: string }) => {
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={value.export} comments={value.comments}>
        {enumType(value)} Enum: {moduleName}.{declName}
      </PanelHeader>
      <EnumSnippet value={value} />
    </div>
  )
}
