import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { MetadataCalls, Ref } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

export const getCalls = (module: Module) => {
  const verbCalls: Ref[] = []

  const metadata = module.verbs
    .map((v) => v.verb)
    .map((v) => v?.metadata)
    .flat()

  const metadataCalls = metadata
    .filter((metadata) => metadata?.value.case === 'calls')
    .map((metadata) => metadata?.value.value as MetadataCalls)

  const calls = metadataCalls.map((metadata) => metadata?.calls).flat()

  calls.forEach((call) => {
    if (!verbCalls.find((v) => v.name === call.name && v.module === call.module)) {
      verbCalls.push({ name: call.name, module: call.module } as Ref)
    }
  })
  return calls
}
