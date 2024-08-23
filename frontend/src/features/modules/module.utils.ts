import {
  ArrowDownOnSquareStackIcon,
  BellAlertIcon,
  BoltIcon,
  CircleStackIcon,
  CodeBracketIcon,
  CogIcon,
  LockClosedIcon,
  RectangleGroupIcon,
} from '@heroicons/react/24/outline'
import type { ForwardRefExoticComponent, SVGProps } from 'react'
import type { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import type { PullSchemaResponse } from '../../protos/xyz/block/ftl/v1/ftl_pb'
import type { MetadataCalls, Ref } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { verbCalls } from '../verbs/verb.utils'

interface InCall {
  module: string
  verb?: string
}

export const getCalls = (module: Module) => {
  const verbCalls: Ref[] = []

  const metadata = module.verbs.map((v) => v.verb).flatMap((v) => v?.metadata)

  const metadataCalls = metadata.filter((metadata) => metadata?.value.case === 'calls').map((metadata) => metadata?.value.value as MetadataCalls)

  const calls = metadataCalls.flatMap((metadata) => metadata?.calls)

  for (const call of calls) {
    if (!verbCalls.find((v) => v.name === call.name && v.module === call.module)) {
      verbCalls.push({ name: call.name, module: call.module } as Ref)
    }
  }

  return calls
}

export const callsIn = (modules: Module[], module: Module) => {
  const allCalls: InCall[] = []
  for (const m of modules) {
    for (const v of m.verbs) {
      const calls = verbCalls(v)
      if (!calls) {
        continue
      }
      for (const call of calls) {
        for (const c of call.calls) {
          if (c.module === module.name) {
            allCalls.push({ module: m.name, verb: v.verb?.name })
          }
        }
      }
    }
  }

  return allCalls
}

export const callsOut = (module: Module) => module.verbs?.flatMap((v) => verbCalls(v))

export const deploymentKeyModuleName = (deploymentKey: string) => {
  const lastIndex = deploymentKey.lastIndexOf('-')
  if (lastIndex !== -1) {
    return deploymentKey.substring(0, lastIndex).replaceAll('dpl-', '')
  }
  return null
}

interface ModuleTreeChild {
  name: string
  href: string
  icon: ForwardRefExoticComponent<SVGProps<SVGSVGElement> & { title?: string; titleId?: string }>
  current?: boolean
}

export interface ModuleTreeItem {
  name: string
  href?: string
  icon: ForwardRefExoticComponent<SVGProps<SVGSVGElement> & { title?: string; titleId?: string }>
  current: boolean
  expanded?: boolean
  children?: ModuleTreeChild[]
}

export const moduleTreeFromSchema = (schema: PullSchemaResponse[]) => {
  const tree: ModuleTreeItem[] = []

  for (const module of schema) {
    tree.push({
      name: module.moduleName,
      icon: RectangleGroupIcon,
      current: false,
      expanded: module.moduleName === 'echo',
      children: [
        { name: 'Data', href: '#', icon: CodeBracketIcon },
        { name: 'Verb', href: '#', icon: BoltIcon },
        { name: 'Database', href: '#', icon: CircleStackIcon },
        { name: 'Config', href: '#', icon: CogIcon },
        { name: 'Secret', href: '#', icon: LockClosedIcon },
        { name: 'FSM', href: '#', icon: ArrowDownOnSquareStackIcon },
        { name: 'Pubsub', href: '#', icon: BellAlertIcon },
      ],
    })
  }

  return tree
}
