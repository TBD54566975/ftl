import { CellsIcon, PackageIcon } from 'hugeicons-react'
import { Config, Data, Database, Enum, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import type { FTLNode } from '../graph/GraphPane'
import { RightPanelHeader } from '../modules/decls/RightPanelHeader'
import { declIcon } from '../modules/module.utils'

export const headerForNode = (node: FTLNode | null) => {
  if (!node) {
    return header({
      IconComponent: CellsIcon,
      content: <div className='text-sm font-medium truncate'>Root</div>,
    })
  }
  if (node instanceof Module) {
    return moduleHeader(node)
  }
  if (node instanceof Verb) {
    if (!node.verb) return
    return <RightPanelHeader Icon={declIcon('verb', node.verb)} title={node.verb.name} />
  }
  if (node instanceof Config) {
    if (!node.config) return
    return <RightPanelHeader Icon={declIcon('config', node.config)} title={node.config.name} />
  }
  if (node instanceof Secret) {
    if (!node.secret) return
    return <RightPanelHeader Icon={declIcon('secret', node.secret)} title={node.secret.name} />
  }
  if (node instanceof Data) {
    if (!node.data) return
    return <RightPanelHeader Icon={declIcon('data', node.data)} title={node.data.name} />
  }
  if (node instanceof Database) {
    if (!node.database) return
    return <RightPanelHeader Icon={declIcon('database', node.database)} title={node.database.name} />
  }
  if (node instanceof Enum) {
    if (!node.enum) return
    return <RightPanelHeader Icon={declIcon('enum', node.enum)} title={node.enum.name} />
  }
}

const header = ({ IconComponent, content }: { IconComponent: React.ElementType; content: React.ReactNode }) => {
  return (
    <div className='flex items-center gap-2 px-2 py-2'>
      <IconComponent className='h-5 w-5 text-indigo-600' />
      <div className='flex flex-col min-w-0'>{content}</div>
    </div>
  )
}

const moduleHeader = (module: Module) => {
  return header({
    IconComponent: PackageIcon,
    content: (
      <>
        <div className='text-sm font-medium truncate'>{module.name}</div>
        <div className='flex pt-1'>
          <span className='font-roboto-mono inline-flex items-center rounded truncate bg-slate-200 dark:bg-slate-700 text-slate-500 dark:text-slate-400 text-xs px-1'>
            <span className='truncate'>{module.deploymentKey}</span>
          </span>
        </div>
      </>
    ),
  })
}
