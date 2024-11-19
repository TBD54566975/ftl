import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Subscription } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import type { ExpandablePanelProps } from '../../../console/ExpandablePanel'

export const subscriptionPanels = (subscription: Subscription) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [<RightPanelAttribute key='name' name='Name' value={subscription.subscription?.name} />],
    },
  ] as ExpandablePanelProps[]
}
