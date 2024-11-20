import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Enum } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import type { ExpandablePanelProps } from '../../../console/ExpandablePanel'

export const enumPanels = (enumDecl: Enum) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={enumDecl.enum?.name} />,
        <RightPanelAttribute key='type' name='Type' value={enumDecl.enum?.type?.value.case} />,
      ],
    },
  ] as ExpandablePanelProps[]
}