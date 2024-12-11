import {
  AnonymousIcon,
  BubbleChatIcon,
  Clock01Icon,
  CodeIcon,
  CodeSquareIcon,
  DatabaseIcon,
  FunctionIcon,
  type HugeiconsProps,
  InternetIcon,
  LeftToRightListNumberIcon,
  MessageIncoming02Icon,
  MessageProgrammingIcon,
  Settings02Icon,
  SquareLock02Icon,
} from 'hugeicons-react'
import type { Module } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import type { Config, Data, Database, Decl, Enum, Secret, Topic, TypeAlias, Verb } from '../../protos/xyz/block/ftl/schema/v1/schema_pb'
import type { MetadataCalls, Ref } from '../../protos/xyz/block/ftl/schema/v1/schema_pb'
import { verbCalls } from './decls/verb/verb.utils'

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

export type DeclSumType = Config | Data | Database | Enum | Topic | TypeAlias | Secret | Verb

export interface DeclInfo {
  declType: string
  value: DeclSumType
  decl: Decl
  path: string
}

export interface ModuleTreeItem {
  name: string
  deploymentKey: string
  decls: DeclInfo[]
  isBuiltin: boolean
}

export const moduleTreeFromStream = (modules: Module[]) => {
  return modules.map(
    (module) =>
      ({
        name: module.name,
        deploymentKey: module.deploymentKey,
        isBuiltin: module.name === 'builtin',
        decls: [
          ...module.configs.map((d) => ({ declType: 'config', value: d.config, decl: d, path: `/modules/${module.name}/config/${d.config?.name}` })),
          ...module.secrets.map((d) => ({ declType: 'secret', value: d.secret, decl: d, path: `/modules/${module.name}/secret/${d.secret?.name}` })),
          ...module.databases.map((d) => ({ declType: 'database', value: d.database, decl: d, path: `/modules/${module.name}/database/${d.database?.name}` })),
          ...module.topics.map((d) => ({ declType: 'topic', value: d.topic, decl: d, path: `/modules/${module.name}/topic/${d.topic?.name}` })),
          ...module.typealiases.map((d) => ({
            declType: 'typealias',
            value: d.typealias,
            decl: d,
            path: `/modules/${module.name}/typealias/${d.typealias?.name}`,
          })),
          ...module.enums.map((d) => ({ declType: 'enum', value: d.enum, decl: d, path: `/modules/${module.name}/enum/${d.enum?.name}` })),
          ...module.data.map((d) => ({ declType: 'data', value: d.data, decl: d, path: `/modules/${module.name}/data/${d.data?.name}` })),
          ...module.verbs.map((d) => ({
            declType: 'verb',
            value: d.verb,
            decl: { value: { case: 'verb', value: d.verb } },
            path: `/modules/${module.name}/verb/${d.verb?.name}`,
          })),
        ],
      }) as ModuleTreeItem,
  )
}

type WithExport = { export?: boolean }

export const declSumTypeIsExported = (d: DeclSumType) => {
  return (d as WithExport).export === true
}

export const declFromModules = (moduleName: string, declCase: string, declName: string, modules?: Module[]) => {
  if (!modules) {
    return undefined
  }
  const module = modules.find((m) => m.name === moduleName)
  if (!module) {
    return undefined
  }
  switch (declCase) {
    case 'config':
      return module.configs.find((d) => d.config?.name === declName)
    case 'data':
      return module.data.find((d) => d.data?.name === declName)
    case 'database':
      return module.databases.find((d) => d.database?.name === declName)
    case 'enum':
      return module.enums.find((d) => d.enum?.name === declName)
    case 'secret':
      return module.secrets.find((d) => d.secret?.name === declName)
    case 'topic':
      return module.topics.find((d) => d.topic?.name === declName)
    case 'typealias':
      return module.typealiases.find((d) => d.typealias?.name === declName)
    case 'verb':
      return module.verbs.find((d) => d.verb?.name === declName)
  }
}

export const hasHideUnexportedInLocalStorage = () => localStorage.getItem('tree_hu') !== null

export const getHideUnexportedFromLocalStorage = () => localStorage.getItem('tree_hu') === 'true'

export const setHideUnexportedFromLocalStorage = (val: boolean) => localStorage.setItem('tree_hu', val ? 'true' : 'false')

export const listExpandedModulesFromLocalStorage = () => (localStorage.getItem('tree_m') || '').split(',').filter((s) => s !== '')

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

export const collapseAllModulesInLocalStorage = () => localStorage.setItem('tree_m', '')

export const declTypeName = (declCase: string, decl: DeclSumType) => {
  const normalizedDeclCase = declCase?.toLowerCase()
  if (normalizedDeclCase === 'verb') {
    const vt = verbTypeFromMetadata(decl as Verb)
    if (vt) {
      return vt
    }
  }
  return normalizedDeclCase || ''
}

const declIcons: Record<string, React.FC<Omit<HugeiconsProps, 'ref'> & React.RefAttributes<SVGSVGElement>>> = {
  config: Settings02Icon,
  data: CodeSquareIcon,
  database: DatabaseIcon,
  enum: LeftToRightListNumberIcon,
  topic: BubbleChatIcon,
  typealias: AnonymousIcon,
  secret: SquareLock02Icon,
  subscription: MessageIncoming02Icon,
  verb: FunctionIcon,
}

export const declIcon = (declCase: string, decl: DeclSumType) => {
  const normalizedDeclCase = declCase?.toLowerCase()

  // Verbs have subtypes as defined by metadata
  const maybeVerbIcon = verbIcon(normalizedDeclCase, decl)
  if (maybeVerbIcon) {
    return maybeVerbIcon
  }

  if (!normalizedDeclCase || !declIcons[normalizedDeclCase]) {
    console.warn(`No icon for decl case: ${declCase}`)
    return CodeIcon
  }

  return declIcons[normalizedDeclCase]
}

const verbIcons: Record<string, React.FC<Omit<HugeiconsProps, 'ref'> & React.RefAttributes<SVGSVGElement>>> = {
  cronjob: Clock01Icon,
  ingress: InternetIcon,
  subscriber: MessageProgrammingIcon,
}

const verbIcon = (declCase: string, decl: DeclSumType) => {
  if (declCase !== 'verb') {
    return
  }
  const vt = verbTypeFromMetadata(decl as Verb)
  if (!vt || !verbIcons[vt]) {
    return declIcons.verb
  }

  return verbIcons[vt]
}

// Most metadata is not mutually exclusive, but schema validation guarantees
// that the ones in this list are.
const verbTypesFromMetadata = ['cronjob', 'ingress', 'subscriber']

export const verbTypeFromMetadata = (verb: Verb) => {
  const found = verb.metadata?.find((m) => m.value.case && verbTypesFromMetadata.includes(m.value.case.toLowerCase()))
  return found?.value.case?.toLowerCase()
}

export const declUrl = (moduleName: string, decl: Decl) => {
  if (!decl?.value?.case) {
    return ''
  }
  return `/modules/${moduleName}/${decl.value.case.toLowerCase()}/${decl.value.value?.name}`
}

export const declUrlFromInfo = (moduleName: string, decl: DeclInfo) => `/modules/${moduleName}/${decl.declType}/${decl.value.name}`

const treeWidthStorageKey = 'tree_w'

export const getTreeWidthFromLS = () => Number(localStorage.getItem(treeWidthStorageKey)) || 300

export const setTreeWidthInLS = (newWidth: number) => localStorage.setItem(treeWidthStorageKey, `${newWidth}`)
