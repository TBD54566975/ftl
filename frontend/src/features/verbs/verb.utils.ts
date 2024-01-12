import { VerbRef } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

export const verbRefString = (verb: VerbRef): string => {
  return `${verb.module}.${verb.name}`
}
