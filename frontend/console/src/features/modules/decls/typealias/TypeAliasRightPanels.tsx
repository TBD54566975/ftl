import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { TypeAlias } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import type { ExpandablePanelProps } from '../../../console/ExpandablePanel'

export const typeAliasPanels = (typeAlias: TypeAlias) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={typeAlias.typealias?.name} />,
        <RightPanelAttribute key='export' name='Type' value={typeAlias.typealias?.type?.value.case} />,
      ],
    },
  ] as ExpandablePanelProps[]
}
