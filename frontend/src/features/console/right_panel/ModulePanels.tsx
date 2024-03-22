import { BoltIcon, CodeBracketIcon, Cog6ToothIcon, LockClosedIcon } from '@heroicons/react/24/outline'
import { Module } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { ExpandablePanelProps } from '../ExpandablePanel'
import { CodeBlock } from '../../../components'

export const modulePanels = (module: Module): ExpandablePanelProps[] => {
  const panels = []

  if (module.verbs && module.verbs.length > 0) {
    panels.push({
      icon: BoltIcon,
      title: 'Verbs',
      expanded: true,
      children: module.verbs.map((v) => <div key={v.verb?.name}>{v.verb?.name}</div>),
    })
  }

  if (module.secrets && module.secrets.length > 0) {
    panels.push({
      icon: LockClosedIcon,
      title: 'Secrets',
      expanded: false,
      children: module.secrets.map((v) => <div key={v.secret?.name}>{v.secret?.name}</div>),
    })
  }

  if (module.configs && module.configs.length > 0) {
    panels.push({
      icon: Cog6ToothIcon,
      title: 'Configs',
      expanded: false,
      children: module.configs.map((v) => <div key={v.config?.name}>{v.config?.name}</div>),
    })
  }

  panels.push({
    icon: CodeBracketIcon,
    title: 'Schema',
    expanded: false,
    children: (
      <div className='p-0'>
        <CodeBlock code={module.schema} language='json' />
      </div>
    ),
    padding: 'p-0',
  })

  return panels
}
