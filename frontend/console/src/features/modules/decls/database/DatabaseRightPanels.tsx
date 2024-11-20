import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Database } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import type { ExpandablePanelProps } from '../../../console/ExpandablePanel'

export const databasePanels = (database: Database) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={database.database?.name} />,
        <RightPanelAttribute key='type' name='Type' value={database.database?.type} />,
      ],
    },
  ] as ExpandablePanelProps[]
}
