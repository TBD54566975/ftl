import { CodeBlock } from '../../../components'
import type { Verb } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../ExpandablePanel'

export const verbPanels = (verb: Verb) => {
  return [
    {
      title: 'Schema',
      expanded: true,
      children: verb?.verb?.response?.toJsonString() && <CodeBlock code={verb?.schema} language='json' />,
      padding: 'p-0',
    },
  ] as ExpandablePanelProps[]
}
