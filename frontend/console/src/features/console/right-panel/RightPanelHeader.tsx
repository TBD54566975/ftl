import { CellsIcon, FunctionIcon, PackageIcon, Settings02Icon, SquareLock02Icon } from 'hugeicons-react'
import { Config, Module, Secret, Verb } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import type { FTLNode } from '../../graph/GraphPane'

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
    return verbHeader(node)
  }
  if (node instanceof Secret) {
    return secretHeader(node)
  }
  if (node instanceof Config) {
    return configHeader(node)
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

const verbHeader = (verb: Verb) => {
  return header({
    IconComponent: FunctionIcon,
    content: (
      <>
        <div className='text-sm font-medium truncate'>{verb.verb?.name}</div>
        <div className='flex'>
          <span className='font-roboto-mono inline-flex items-center rounded truncate bg-slate-200 dark:bg-slate-800 text-slate-500 dark:text-slate-400 text-xs'>
            <span className='truncate'>{verb.verb?.name}</span>
          </span>
        </div>
      </>
    ),
  })
}

const secretHeader = (secret: Secret) => {
  return header({
    IconComponent: SquareLock02Icon,
    content: (
      <>
        <div className='text-sm font-medium truncate'>Secret</div>
        <div className='flex'>
          <span className='font-roboto-mono inline-flex items-center rounded truncate bg-slate-200 dark:bg-slate-800 text-slate-500 dark:text-slate-400 text-xs'>
            <span className='truncate'>{secret.secret?.name}</span>
          </span>
        </div>
      </>
    ),
  })
}

const configHeader = (config: Config) => {
  return header({
    IconComponent: Settings02Icon,
    content: (
      <>
        <div className='text-sm font-medium truncate'>{config.config?.name}</div>
        <div className='flex'>
          <span className='font-roboto-mono inline-flex items-center rounded truncate bg-slate-200 dark:bg-slate-800 text-slate-500 dark:text-slate-400 text-xs'>
            <span className='truncate'>{config.config?.type?.value.case}</span>
          </span>
        </div>
      </>
    ),
  })
}
