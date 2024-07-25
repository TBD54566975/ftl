import { CodeBlock } from '../../../components'
import { Verb } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { ExpandablePanelProps } from '../ExpandablePanel'

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
