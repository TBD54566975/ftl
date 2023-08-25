import {
  MetadataCalls,
  Verb,
} from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

export function getCalls(verb?: Verb): MetadataCalls[] {
  return (
    verb?.metadata
      ?.filter(meta => meta.value.case === 'calls')
      .map(meta => meta.value.value as MetadataCalls) ?? []
  )
}

export function buildVerbSchema(
  verbSchema: string,
  dataScemas: string[]
): string {
  return dataScemas.join('\n\n') + '\n\n' + verbSchema
}
