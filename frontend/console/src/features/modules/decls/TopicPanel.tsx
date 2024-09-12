import type { Topic } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { PanelHeader } from './PanelHeader'
import { TypeEl } from './TypeEl'

export const TopicPanel = ({ value, moduleName, declName }: { value: Topic; moduleName: string; declName: string }) => {
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={value.export} comments={value.comments}>
        Topic: {moduleName}.{declName}
      </PanelHeader>
      <div className='text-sm my-4'>
        Event Type: <TypeEl t={value.event} />
      </div>
    </div>
  )
}
