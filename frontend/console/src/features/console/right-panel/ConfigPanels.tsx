import { RightPanelAttribute } from '../../../components/RightPanelAttribute'
import type { Config } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../ExpandablePanel'

export const configPanels = (config: Config) => {
  return [
    {
      title: config.config?.name,
      expanded: true,
      children: <RightPanelAttribute name='Name' value={config.config?.name} />,
    },
  ] as ExpandablePanelProps[]
}
