import {
  AnonymousIcon,
  BubbleChatIcon,
  CodeIcon,
  DatabaseIcon,
  FlowIcon,
  FunctionIcon,
  type HugeiconsProps,
  LeftToRightListNumberIcon,
  MessageIncoming02Icon,
  Settings02Icon,
  SquareLock02Icon,
} from 'hugeicons-react'
import type { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import type { PullSchemaResponse } from '../../protos/xyz/block/ftl/v1/ftl_pb'
import type { Decl } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
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

export interface ModuleTreeItem {
  name: string
  deploymentKey: string
  decls: Decl[]
  isBuiltin: boolean
}

export const moduleTreeFromSchema = (schema: PullSchemaResponse[]) => {
  const tree = schema.map((module) => ({
    name: module.moduleName,
    deploymentKey: module.deploymentKey,
    isBuiltin: module.moduleName === 'builtin',
    decls: module.schema ? module.schema.decls : [],
  }))
  return tree
}

export const declFromSchema = (moduleName: string, declName: string, schema: PullSchemaResponse[]) => {
  const module = schema.find((m) => m.moduleName === moduleName)
  if (!module?.schema) {
    return undefined
  }
  return module.schema.decls.find((d) => d.value.value?.name === declName)
}

export const listExpandedModulesFromLocalStorage = () => (localStorage.getItem('tree_m') || '').split(',')

export const toggleModuleExpansionInLocalStorage = (moduleName: string) => {
  const expanded = listExpandedModulesFromLocalStorage()
  const i = expanded.indexOf(moduleName)
  if (i === -1) {
    localStorage.setItem('tree_m', [...expanded, moduleName].join(','))
  } else {
    expanded.splice(i, 1)
    localStorage.setItem('tree_m', expanded.join(','))
  }
}

export const addModuleToLocalStorageIfMissing = (moduleName?: string) => {
  const expanded = listExpandedModulesFromLocalStorage()
  if (moduleName && !expanded.includes(moduleName)) {
    localStorage.setItem('tree_m', [...expanded, moduleName].join(','))
  }
}

type IconMap = Record<string, React.FC<Omit<HugeiconsProps, 'ref'> & React.RefAttributes<SVGSVGElement>>>
export const declIcons: IconMap = {
  config: Settings02Icon,
  data: CodeIcon,
  database: DatabaseIcon,
  enum: LeftToRightListNumberIcon,
  fsm: FlowIcon,
  topic: BubbleChatIcon,
  typeAlias: AnonymousIcon,
  secret: SquareLock02Icon,
  subscription: MessageIncoming02Icon,
  verb: FunctionIcon,
}

export const declUrl = (moduleName: string, decl: Decl) => `/modules/${moduleName}/${decl.value.case}/${decl.value.value?.name}`
