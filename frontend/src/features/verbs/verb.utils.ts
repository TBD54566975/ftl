import { Ref } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

export const verbRefString = (verb: Ref): string => {
  return `${verb.module}.${verb.name}`
}
