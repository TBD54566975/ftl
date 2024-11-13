import type { Subscription } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { Schema } from '../schema/Schema'
import { PanelHeader } from './PanelHeader'
import { References } from './References'

export const SubscriptionPanel = ({ value, schema, moduleName, declName }: { value: Subscription; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={false} comments={value.subscription?.comments}>
        Subscription: {moduleName}.{declName}
      </PanelHeader>
      <div className='-mx-3.5'>
        <Schema schema={schema} />
      </div>
      <References references={value.references} />
    </div>
  )
}
