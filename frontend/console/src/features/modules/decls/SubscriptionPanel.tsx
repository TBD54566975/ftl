import type { Subscription } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { PanelHeader } from './PanelHeader'
import { DeclLink } from './DeclLink'
import { RefLink, TypeEl } from './TypeEl'

export const SubscriptionPanel = ({ value, moduleName, declName }: { value: Subscription; moduleName: string; declName: string }) => {
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={value.export} comments={value.comments}>
        Subscription: {moduleName}.{declName}
      </PanelHeader>
      <div className='text-sm my-4'>
        To Topic: <RefLink r={value.topic} />
      </div>
    </div>
  )
}
