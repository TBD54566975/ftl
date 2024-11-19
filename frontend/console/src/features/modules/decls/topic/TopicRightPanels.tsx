import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Topic } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import type { ExpandablePanelProps } from '../../../console/ExpandablePanel'

export const topicPanels = (topic: Topic) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [<RightPanelAttribute key='name' name='Name' value={topic.topic?.name} />],
    },
  ] as ExpandablePanelProps[]
}
