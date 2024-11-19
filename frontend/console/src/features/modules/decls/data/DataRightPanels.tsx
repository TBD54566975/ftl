import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Data } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import type { ExpandablePanelProps } from '../../../console/ExpandablePanel'

export const dataPanels = (data: Data) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [<RightPanelAttribute key='name' name='Name' value={data.data?.name} />],
    },
  ] as ExpandablePanelProps[]
}
