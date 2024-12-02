import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Database } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../graph/ExpandablePanel'
import { DeclDefaultPanels } from '../DeclDefaultPanels'

export const databasePanels = (database: Database, schema?: string) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={database.database?.name} />,
        <RightPanelAttribute key='type' name='Type' value={database.database?.type} />,
      ],
    },
    ...DeclDefaultPanels(schema, database.references),
  ] as ExpandablePanelProps[]
}
