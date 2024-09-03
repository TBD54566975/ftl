import type { Config } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { PanelHeader } from './PanelHeader'
import { TypeEl } from './TypeEl'

export const ConfigPanel = ({ value, moduleName, declName }: { value: Config; moduleName: string; declName: string }) => {
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={false} comments={value.comments}>
        <p>
          Config: {moduleName}.{declName}
        </p>
      </PanelHeader>
      <div className='text-sm my-4'>
        Type: <TypeEl t={value.type} />
      </div>
    </div>
  )
}
