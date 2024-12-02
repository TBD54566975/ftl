import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Topic } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../graph/ExpandablePanel'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
export const topicPanels = (topic: Topic, schema?: string) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={topic.topic?.name} />,
        <RightPanelAttribute key='export' name='Event' value={topic.topic?.event?.value.case} />,
      ],
    },
    ...DeclDefaultPanels(schema, topic.references),
  ] as ExpandablePanelProps[]
}
