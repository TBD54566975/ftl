import type { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
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
