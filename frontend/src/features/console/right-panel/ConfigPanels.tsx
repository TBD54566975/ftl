import { RightPanelAttribute } from '../../../components/RightPanelAttribute'
import { Config } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { ExpandablePanelProps } from '../ExpandablePanel'

export const configPanels = (config: Config) => {
  return [
    {
      title: config.config?.name,
      expanded: true,
      children: <RightPanelAttribute name='Name' value={config.config?.name} />,
    },
  ] as ExpandablePanelProps[]
}
