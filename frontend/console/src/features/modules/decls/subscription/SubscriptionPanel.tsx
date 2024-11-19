import { ResizablePanels } from '../../../../components/ResizablePanels'
import type { Subscription } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import { declIcon } from '../../module.utils'
import { Schema } from '../../schema/Schema'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { subscriptionPanels } from './SubscriptionRightPanels'

export const SubscriptionPanel = ({ value, schema, moduleName, declName }: { value: Subscription; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }

  const decl = value.subscription
  if (!decl) {
    return
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <div className=''>
              <PanelHeader title='Data' declRef={`${moduleName}.${declName}`} exported={false} comments={decl.comments} />
              <div className='-mx-3.5'>
                <Schema schema={schema} />
              </div>
            </div>
          </div>
        }
        rightPanelHeader={<RightPanelHeader Icon={declIcon('subscription', decl)} title={declName} />}
        rightPanelPanels={[...subscriptionPanels(value), ...DeclDefaultPanels(schema, value.references)]}
        storageKeyPrefix='subscriptionPanel'
      />
    </div>
  )
}
