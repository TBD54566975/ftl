import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Secret } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import type { ExpandablePanelProps } from '../../../console/ExpandablePanel'

export const secretPanels = (secret: Secret) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={secret.secret?.name} />,
        <RightPanelAttribute key='type' name='Type' value={secret.secret?.type?.value.case ?? ''} />,
      ],
    },
  ] as ExpandablePanelProps[]
}
