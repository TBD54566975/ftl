import type { TypeAlias } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { PanelHeader } from './PanelHeader'
import { TypeEl } from './TypeEl'

export const TypeAliasPanel = ({ value, moduleName, declName }: { value: TypeAlias; moduleName: string; declName: string }) => {
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={value.export} comments={value.comments}>
        <p>
          Type Alias: {moduleName}.{declName}
        </p>
      </PanelHeader>
      <div className='text-sm my-4'>
        Underlying type: <TypeEl t={value.type} />
      </div>
    </div>
  )
}
