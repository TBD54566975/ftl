import { RightPanelAttribute } from '../../../components/RightPanelAttribute'
import type { Secret } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../ExpandablePanel'

export const secretPanels = (secret: Secret) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: <RightPanelAttribute name='Type' value={secret.secret?.type?.value?.case} />,
    },
  ] as ExpandablePanelProps[]
}
