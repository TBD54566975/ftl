import type { Database } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { PanelHeader } from './PanelHeader'

export const DatabasePanel = ({ value, moduleName, declName }: { value: Database; moduleName: string; declName: string }) => {
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={false} comments={value.comments}>
        <p>
          Database: {moduleName}.{declName}
        </p>
      </PanelHeader>
      <div className='text-sm my-4'>Type: {value.type}</div>
    </div>
  )
}
