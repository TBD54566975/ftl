import { BoltIcon, Cog6ToothIcon, CubeIcon, LockClosedIcon, RectangleGroupIcon } from '@heroicons/react/24/outline'
import { Config, Module, Secret, Verb } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { FTLNode } from '../../graph/GraphPage'

export const headerForNode = (node: FTLNode | null) => {
  if (!node) {
    return header({
      IconComponent: CubeIcon,
      content: <div className='text-sm font-medium truncate'>Root</div>,
    })
  }
  if (node instanceof Module) {
    return moduleHeader(node)
  } else if (node instanceof Verb) {
    return verbHeader(node)
  } else if (node instanceof Secret) {
    return secretHeader(node)
  } else if (node instanceof Config) {
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
    IconComponent: RectangleGroupIcon,
    content: (
      <>
        <div className='text-sm font-medium truncate'>{module.name}</div>
        <div className='flex'>
          <span className='font-roboto-mono inline-flex items-center rounded truncate bg-slate-200 dark:bg-slate-800 text-slate-500 dark:text-slate-400 text-xs'>
            <span className='truncate'>{module.deploymentKey}</span>
          </span>
        </div>
      </>
    ),
  })
}

const verbHeader = (verb: Verb) => {
  return header({
    IconComponent: BoltIcon,
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
    IconComponent: LockClosedIcon,
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
    IconComponent: Cog6ToothIcon,
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
