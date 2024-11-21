import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Subscription } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import type { ExpandablePanelProps } from '../../../console/ExpandablePanel'
import { DeclDefaultPanels } from '../DeclDefaultPanels'

export const subscriptionPanels = (subscription: Subscription, schema?: string) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={subscription.subscription?.name} />,
        <RightPanelAttribute key='topic' name='Topic' value={subscription.subscription?.topic?.name} />,
      ],
    },
    ...DeclDefaultPanels(schema, subscription.references),
  ] as ExpandablePanelProps[]
}
