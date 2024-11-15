import type { Ref } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import type { ExpandablePanelProps } from '../../console/ExpandablePanel'
import { Schema } from '../schema/Schema'
import { References } from './References'

export const DeclDefaultPanels = (schema: string, references: Ref[]) => {
  const panels = [] as ExpandablePanelProps[]

  panels.push({
    title: 'Schema',
    expanded: true,
    children: <Schema schema={schema} />,
  })

  panels.push({
    title: 'References',
    expanded: true,
    children: <References references={references} />,
  })

  return panels
}
