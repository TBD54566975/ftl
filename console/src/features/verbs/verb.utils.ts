import { MetadataCalls, Verb } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

export function getCalls(verb: Verb): MetadataCalls[] {
  return (
    verb?.metadata?.filter(meta => meta.value.case === 'calls').map(meta => meta.value.value as MetadataCalls) ?? []
  )
}

export function getVerbCode(verb: Verb): string {
  let codeBlock = verb.comments.map(comment => `// ${comment}`).join('\n')
  if (verb.comments.length > 0) codeBlock += '\n'
  codeBlock += `${verb.name}(${verb.request?.name}) -> ${verb.response?.name}`

  return codeBlock
}
