import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Config } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import type { ExpandablePanelProps } from '../../../console/ExpandablePanel'

export const configPanels = (config: Config) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={config.config?.name} />,
        <RightPanelAttribute key='type' name='Type' value={config.config?.type?.value.case ?? ''} />,
      ],
    },
  ] as ExpandablePanelProps[]
}
